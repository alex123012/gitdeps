package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/alex123012/gitdeps/cmd/common"
	"github.com/alex123012/gitdeps/pkg/gitlab"
	"github.com/alex123012/gitdeps/pkg/webhook"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	port   int
	codecs serializer.CodecFactory = serializer.NewCodecFactory(runtime.NewScheme())
	logger *log.Logger             = hclog.L().StandardLogger(&hclog.StandardLoggerOptions{
		InferLevels:              true,
		InferLevelsWithTimestamp: true,
	})
)

func NewWebHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-handler",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			// fromFile, err := cmd.Parent().Flags().GetBool("from-file")
			// if err != nil {
			// 	return err
			// }
			// flags := cmd.InheritedFlags()
			// flags.PrintDefaults()
			// fmt.Println(fromFile)
			return RunWebHook(ctx, fromFile)
		},
	}

	cmd.Flags().IntVar(&port, "port", port, "port to expose")
	return cmd
}

// https://github.com/cnych/admission-webhook-example/blob/master/webhook.go
func RunWebHook(ctx context.Context, fromFile bool) error {
	fmt.Println("Starting webhook server")
	var cert tls.Certificate
	var err error
	if fromFile {
		cfg := &common.Config.WebhookConf
		certsPath := cfg.Tls.Path
		certificateFile := path.Join(certsPath, cfg.Tls.CertFile)
		keyFile := path.Join(certsPath, cfg.Tls.KeyFile)

		cert, err = tls.LoadX509KeyPair(certificateFile, keyFile)
	} else {
		caMap, err := webhook.GenerateCertificate(ctx, common.Config.WebhookConf, fromFile)
		if err != nil {
			return err
		}
		cert, err = tls.X509KeyPair(caMap["cert"].Bytes(), caMap["key"].Bytes())
	}

	if err != nil {
		return err
	}

	if port == 0 {
		port = int(*common.Config.WebhookConf.Webhook.ClientConfig.Service.Port)
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", ValidateDeployingBranch)
	server := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		ErrorLog: logger,
		Handler:  mux,
	}
	go func() {
		// listening OS shutdown singal
		<-ctx.Done()
		logger.Printf("Got OS shutdown signal, shutting down webhook server gracefully...")
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServeTLS("", "")
}

func ValidateDeployingBranch(w http.ResponseWriter, r *http.Request) {

	deserializer := codecs.UniversalDeserializer()
	admissionReviewRequest, err := GetAdmissionRequest(r, deserializer)
	if err != nil {
		ReturnError(w, 400,
			fmt.Sprintf("error getting admission review from request: %v", err),
		)
		return
	}
	// err = config.PrintStruct(admissionReviewRequest)
	// if err != nil {
	// 	ReturnError(w, 500,
	// 		fmt.Sprintf("error printing admission review: %v", err),
	// 	)
	// 	return
	// }

	rawRequest := admissionReviewRequest.Request.Object.Raw
	object := unstructured.Unstructured{}
	if _, _, err := deserializer.Decode(rawRequest, nil, &object); err != nil {
		ReturnError(w, 500,
			fmt.Sprintf("error decoding raw resource object: %v", err),
		)
		return
	}
	// err = config.PrintStruct(object)
	// if err != nil {
	// 	ReturnError(w, 500,
	// 		fmt.Sprintf("error printing admission review: %v", err),
	// 	)
	// 	return
	// }
	annotations := object.GetAnnotations()
	pipelineUrl := annotations["gitlab.ci.werf.io/pipeline-url"]
	allowValidation, err := gitlab.TargetHaveAllCommitsFromDefault(pipelineUrl)
	fmt.Println(allowValidation)
	if err != nil {
		ReturnError(w, 500,
			fmt.Sprintf("error validating resource from gitlab api: %v", err),
		)

		return
	}

	var message, status string
	var warnings []string
	if allowValidation {
		message = "All good"
		status = "success"
	} else {
		message = "Deploying branch don't have commits from default branch"
		status = "error"
		warnings = []string{message}
	}

	admissionReviewResponse := admissionv1.AdmissionReview{
		Response: &admissionv1.AdmissionResponse{
			Allowed: allowValidation,
			Result: &metav1.Status{
				Message: message,
				Status:  status,
				Reason:  metav1.StatusReasonConflict,
			},
			Warnings: warnings,
			UID:      admissionReviewRequest.Request.UID,
		},
	}
	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())

	resp, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		ReturnError(w, 500,
			fmt.Sprintf("error marshalling response json: %v", err),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func GetAdmissionRequest(r *http.Request, deserializer runtime.Decoder) (*admissionv1.AdmissionReview, error) {
	// Validate that the incoming content type is correct.
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("expected application/json content-type")
	}
	// Get the body data, which will be the AdmissionReview
	// content for the request.
	var body []byte
	defer r.Body.Close()
	if r.Body != nil {
		requestData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = requestData
	}

	// Decode the request body into
	admissionReviewRequest := &admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, admissionReviewRequest); err != nil {
		return nil, err
	}

	return admissionReviewRequest, nil
}

func ReturnError(w http.ResponseWriter, status int, msg string) {
	// msg := fmt.Sprintf("error getting admission review from request: %v", err)
	hclog.L().Error(msg)
	w.WriteHeader(400)
	w.Write([]byte(msg))
}
