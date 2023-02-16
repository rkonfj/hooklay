package internalutil

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func ParseCommitIDFromImage(image string) string {
	nameAndVersion := strings.Split(image, ":")
	if len(nameAndVersion) != 2 {
		logrus.Error("image format error")
		return ""
	}
	versionDetails := strings.Split(nameAndVersion[1], "-")
	if len(versionDetails) != 3 {
		logrus.Error("image format error")
		return ""
	}
	return versionDetails[1]
}
