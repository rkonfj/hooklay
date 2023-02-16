package notification

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/sirupsen/logrus"
)

func SendDingtalk(title, markdown string) error {
	var body bytes.Buffer
	body.WriteString(fmt.Sprintf(`{
        "msgtype": "markdown",
        "markdown": {
          "title": "%s",
          "text": "# %s\n---\n%s"
        }
    }`, title, title, markdown))
	resp, err := http.Post(config.Conf.Notification.Dingtalk, "application/json;charset=utf-8", &body)
	if err != nil {
		return err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logrus.Info(string(bodyBytes))
	return nil
}
