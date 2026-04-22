package module

import (
	"errors"
	"fmt"
	ref "github.com/distribution/reference"
	"github.com/onlyLTY/dockerCopilot/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
)

// ImageCheckList 检查更新处理后的镜像列表
type ImageCheckList struct {
	NeedUpdate bool
}
type ImageUpdateData struct {
	Data       map[string]ImageCheckList
	httpClient *http.Client
}

const ContentDigestHeader = "Docker-Content-Digest"

func NewImageCheck(httpClient *http.Client) *ImageUpdateData {
	return &ImageUpdateData{
		Data:       map[string]ImageCheckList{},
		httpClient: httpClient,
	}
}
func (i *ImageUpdateData) CheckUpdate(imageList []types.Image) {
	for _, image := range imageList {
		if strings.Contains(image.ImageName, "0nlylty/dockercopilot") {
			continue
		}
		i.checkSingleImage(image)
	}
}

func (i *ImageUpdateData) checkSingleImage(image types.Image) {
	imageKey := image.ImageName + ":" + image.ImageTag
	token, err := GetToken(image, "", i.httpClient)
	if err != nil {
		logx.Errorf("[%s] 获取token失败，继续尝试: %s", imageKey, err.Error())
	}
	digestURL, err := BuildManifestURL(image, i.httpClient)
	if err != nil {
		logx.Errorf("[%s] 构建manifest URL失败: %s", imageKey, err.Error())
		return
	}
	remoteDigest, err := GetDigest(digestURL, token, i.httpClient)
	if err != nil {
		logx.Errorf("[%s] 获取远端digest失败: %s", imageKey, err.Error())
		return
	}
	if len(image.RepoDigests) == 0 {
		logx.Errorf("[%s] 本地无repoDigest，跳过", imageKey)
		return
	}
	if remoteDigest == "" {
		logx.Errorf("[%s] 远端digest为空，跳过", imageKey)
		return
	}

	needUpdate := true
	for _, localRepoDigest := range image.RepoDigests {
		parts := strings.SplitN(localRepoDigest, "@", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[1] == "" {
			continue
		}
		if remoteDigest == parts[1] {
			needUpdate = false
			break
		}
	}
	if needUpdate {
		logx.Infof("[%s] 有更新可用 (remoteDigest: %s)", imageKey, remoteDigest)
	} else {
		logx.Infof("[%s] 已是最新", imageKey)
	}
	i.Data[image.ID] = ImageCheckList{NeedUpdate: needUpdate}
}

func BuildManifestURL(image types.Image, httpClient *http.Client) (string, error) {
	normalizedRef, err := ref.ParseDockerRef(image.ImageName + ":" + image.ImageTag)
	if err != nil {
		return "", err
	}
	normalizedTaggedRef, isTagged := normalizedRef.(ref.NamedTagged)
	if !isTagged {
		return "", errors.New("镜像无tag" + normalizedRef.String())
	}

	host, ErrGetRegistryAddress := GetRegistryAddress(normalizedTaggedRef.Name(), httpClient)
	img, tag := ref.Path(normalizedTaggedRef), normalizedTaggedRef.Tag()

	if ErrGetRegistryAddress != nil {
		return "", ErrGetRegistryAddress
	}

	url := url2.URL{
		Scheme: "https",
		Host:   host,
		Path:   fmt.Sprintf("/v2/%s/manifests/%s", img, tag),
	}
	return url.String(), nil
}

func GetDigest(url string, token string, httpClient *http.Client) (string, error) {
	digest, err := requestDigest("HEAD", url, token, httpClient)
	if err != nil {
		logx.Infof("HEAD 请求获取digest失败，尝试GET: %s", err.Error())
		return requestDigest("GET", url, token, httpClient)
	}
	return digest, nil
}

func requestDigest(method string, url string, token string, httpClient *http.Client) (string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}

	if token != "" {
		req.Header.Add("Authorization", token)
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logx.Error("关闭body失败" + err.Error())
		}
	}(res.Body)

	if method == "GET" {
		io.Copy(io.Discard, res.Body)
	}

	if res.StatusCode != 200 {
		wwwAuthHeader := res.Header.Get("www-authenticate")
		if wwwAuthHeader == "" {
			wwwAuthHeader = "not present"
		}
		return "", fmt.Errorf("%s %s responded with %q, auth: %q", method, url, res.Status, wwwAuthHeader)
	}
	digest := res.Header.Get(ContentDigestHeader)
	if digest == "" {
		return "", fmt.Errorf("%s %s 响应中无 %s header", method, url, ContentDigestHeader)
	}
	return digest, nil
}
