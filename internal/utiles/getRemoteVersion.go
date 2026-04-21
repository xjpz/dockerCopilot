package utiles

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/onlyLTY/dockerCopilot/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetRemoteVersion(httpClient *http.Client) (remoteVersion string, err error) {
	versionURL := buildGitHubURL("https://raw.githubusercontent.com/onlyLTY/dockerCopilot/UGREEN/version")
	remoteVersion, err = fetchVersionFromURL(httpClient, versionURL)
	if err != nil {
		return "0.0.0", err
	}

	localVersion := config.Version
	if strings.Contains(localVersion, "FNOS") {
		logx.Infof("飞牛版本，无需在线更新")
		return localVersion, nil
	}
	if localVersion == remoteVersion {
		logx.Info("版本一致:", localVersion)
		return remoteVersion, nil
	} else {
		logx.Infof("版本不一致! 本地: %s, 远程: %s\n", localVersion, remoteVersion)
		return remoteVersion, nil
	}
}

func buildGitHubURL(rawURL string) string {
	githubProxy := os.Getenv("githubProxy")
	if githubProxy != "" {
		githubProxy = strings.TrimRight(githubProxy, "/") + "/"
		return githubProxy + rawURL
	}
	return rawURL
}

func fetchVersionFromURL(httpClient *http.Client, url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logx.Error("关闭Body失败:", err)
		}
	}(resp.Body)

	versionData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(versionData)), nil
}
