package senderscore

import (
	"fmt"
)

var reportPath = "/senderscore/report/?lookup=%s&authenticated=true"

type SenderClient struct {
	request   *RequestWrapper
	userAgent string
}

func NewSenderClient(rw *RequestWrapper) *SenderClient {
	return &SenderClient{
		request:   rw,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	}
}

func (c *SenderClient) GetReport(ip string) (string, error) {
	url := fmt.Sprintf(reportPath, ip)
	htmlContent, err := c.request.SendRequest("GET", url, c.userAgent)
	if err != nil {
		return "", err
	}

	return string(htmlContent), nil
}
