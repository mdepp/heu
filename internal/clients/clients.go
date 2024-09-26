package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/go-resty/resty/v2"
	"github.com/pion/dtls/v3"
)

type HueClipAPIClient struct {
	httpClient *resty.Client
}

type HueClipAPIClientConfig struct {
	BridgeIP          net.IP
	RootCertFilePath  string
	HueApplicationKey string
}

func NewHueClipAPIClient(config *HueClipAPIClientConfig) *HueClipAPIClient {
	httpClient := resty.New().
		SetBaseURL(fmt.Sprintf("https://%s/clip/v2", config.BridgeIP.String())).
		SetRootCertificate(config.RootCertFilePath).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetHeader("hue-application-key", config.HueApplicationKey)

	return &HueClipAPIClient{httpClient: httpClient}
}

func (c *HueClipAPIClient) ListEntertainmentConfigurations() (*EntertainmentConfigurationResult, error) {
	result := &EntertainmentConfigurationResult{}
	response, err := c.httpClient.R().SetResult(result).Get("/resource/entertainment_configuration")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if response.IsError() {
		return nil, fmt.Errorf("response failed: %s", response.Status())
	}
	return result, nil
}

func (c *HueClipAPIClient) GetEntertainmentConfiguration(id string) (*EntertainmentConfigurationResult, error) {
	result := &EntertainmentConfigurationResult{}
	response, err := c.httpClient.R().SetResult(result).Get(fmt.Sprintf("/resource/entertainment_configuration/%s", id))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if response.IsError() {
		return nil, fmt.Errorf("response failed: %s", response.Status())
	}
	return result, nil
}

func (c *HueClipAPIClient) StartEntertainmentSession(id string) error {
	response, err := c.httpClient.R().
		SetBody(map[string]string{"action": "start"}).
		Put(fmt.Sprintf("/resource/entertainment_configuration/%s", id))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if response.IsError() {
		return fmt.Errorf("response failed: %s", response.Status())
	}
	return nil
}

func (c *HueClipAPIClient) StopEntertainmentSession(id string) error {
	response, err := c.httpClient.R().
		SetBody(map[string]string{"action": "stop"}).
		Put(fmt.Sprintf("/resource/entertainment_configuration/%s", id))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if response.IsError() {
		return fmt.Errorf("response failed: %s", response.Status())
	}
	return nil
}

type HueEntertainmentAPIClient struct {
	dtlsConn *dtls.Conn
}

type HueEntertainmentAPIClientConfig struct {
	BridgeIP          net.IP
	HueApplicationKey string
	DTLSClientKey     []byte
}

func NewHueEntertainmentAPIClient(config *HueEntertainmentAPIClientConfig) (*HueEntertainmentAPIClient, error) {
	bridgeUDPAddr := &net.UDPAddr{IP: config.BridgeIP, Port: 2100}
	dtlsConfig := &dtls.Config{
		PSK: func(hint []byte) ([]byte, error) {
			return config.DTLSClientKey, nil
		}, PSKIdentityHint: []byte(config.HueApplicationKey),
		CipherSuites:       []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_GCM_SHA256},
		InsecureSkipVerify: true,
	}
	dtlsConn, err := dtls.Dial("udp4", bridgeUDPAddr, dtlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	return &HueEntertainmentAPIClient{dtlsConn: dtlsConn}, nil
}

func (c *HueEntertainmentAPIClient) Close() error {
	return c.dtlsConn.Close()
}

func (c *HueEntertainmentAPIClient) Handshake(ctx context.Context) error {
	if err := c.dtlsConn.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("failed to handshake with server: %w", err)
	}
	return nil
}

func (c *HueEntertainmentAPIClient) Write(body []byte) (int, error) {
	return c.dtlsConn.Write(body)
}
