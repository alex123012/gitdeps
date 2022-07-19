package webhook

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/alex123012/gitdeps/pkg/config"
	"github.com/hashicorp/go-hclog"
	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1Typed "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"k8s.io/client-go/rest"
)

func GenerateCertificate(WebhookConf config.WebHookConf, toFile bool) (map[string]*bytes.Buffer, error) {
	certsPath := WebhookConf.Tls.Path
	certificateFile := path.Join(certsPath, WebhookConf.Tls.CertFile)
	keyFile := path.Join(certsPath, WebhookConf.Tls.KeyFile)
	organization := WebhookConf.Tls.Organization
	if toFile {
		if _, err := os.Stat(certificateFile); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not valid certificate path: %w", err)
		}
		if _, err := os.Stat(keyFile); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not valid key path: %w", err)
		}
	}

	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err != nil {
		return nil, err
	}
	serviceName := WebhookConf.Webhook.ClientConfig.Service.Name
	serviceNamespace := WebhookConf.Webhook.ClientConfig.Service.Namespace
	nameNamespace := fmt.Sprintf("%s.%s", serviceName, serviceNamespace)

	commonName := fmt.Sprintf("%s.svc", nameNamespace)

	dnsNames := []string{
		serviceName,
		nameNamespace,
		commonName,
	}

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// PEM encode the  server cert and key
	serverCertPEM = new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	if err != nil {
		return nil, err
	}

	serverPrivKeyPEM = new(bytes.Buffer)
	err = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})
	if err != nil {
		return nil, err
	}

	if toFile {
		err = os.MkdirAll(certsPath, 0666)
		if err != nil {
			return nil, err
		}
		err = WriteFile(certificateFile, serverCertPEM)
		if err != nil {
			return nil, err
		}

		err = WriteFile(keyFile, serverPrivKeyPEM)
		if err != nil {
			return nil, err
		}
	}
	return map[string]*bytes.Buffer{
		"ca":   caPEM,
		"cert": serverCertPEM,
		"key":  serverPrivKeyPEM,
	}, nil
}

// WriteFile writes data in the file at the given path
func WriteFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func CreateWebhookConf(ctx context.Context, config *rest.Config, caPEM *bytes.Buffer, WebhookConf config.WebHookConf) error {

	api, err := AdmissionApiFromConfig(config)
	if err != nil {
		return err
	}

	WebhookConf.Webhook.ClientConfig.CABundle = caPEM.Bytes()
	webHookConfResource := &v1.ValidatingWebhookConfiguration{
		ObjectMeta: WebhookConf.Metadata,
		Webhooks:   []v1.ValidatingWebhook{WebhookConf.Webhook},
	}
	hclog.L().Info("Trying to get WebhookConfiguration")
	result, err := api.Get(ctx, webHookConfResource.ObjectMeta.Name, metav1.GetOptions{})
	var check *v1.ValidatingWebhookConfiguration

	if err != nil {
		hclog.L().Info("Creating WebhookConfiguration")
		check, err = api.Create(ctx, webHookConfResource, metav1.CreateOptions{})
	} else {
		hclog.L().Info("Updating WebhookConfiguration")
		webHookConfResource.ObjectMeta = result.ObjectMeta
		check, err = api.Update(ctx, webHookConfResource, metav1.UpdateOptions{})
	}

	if err != nil {
		return err
	}

	if check.ObjectMeta.Name != webHookConfResource.ObjectMeta.Name {
		return fmt.Errorf("something went wrong with creating Validaton Webhook")
	}
	return nil
}

func AdmissionApiFromConfig(config *rest.Config) (v1Typed.ValidatingWebhookConfigurationInterface, error) {
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient.AdmissionregistrationV1().ValidatingWebhookConfigurations(), nil
}
