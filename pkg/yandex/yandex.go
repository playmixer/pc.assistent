package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	URIOAuthToken  = "https://iam.api.cloud.yandex.net/iam/v1/tokens"
	URIRecognizeV1 = "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"
)

type Lang string

const (
	LngAuto Lang = "auto"  //автоматическое распознавание языка
	LngDE   Lang = "de-DE" //немецкий
	LngUS   Lang = "en-US" //английский
	LngES   Lang = "es-ES" //испанский
	LngFI   Lang = "fi-FI" //финский
	LngFR   Lang = "fr-FR" //французский
	LngHE   Lang = "he-HE" //иврит
	LngIT   Lang = "it-IT" //итальянский
	LngKZ   Lang = "kk-KZ" //казахский
	LngNL   Lang = "nl-NL" //голландский
	LngPL   Lang = "pl-PL" //польский
	LngPT   Lang = "pt-PT" //португальский
	LngBR   Lang = "pt-BR" //бразильский португальский
	LngRU   Lang = "ru-RU" //русский язык (по умолчанию)
	LngSE   Lang = "sv-SE" //шведский
	LngTR   Lang = "tr-TR" //турецкий
	LngUZ   Lang = "uz-UZ" //узбекский (латиница)
)

type Format string

const (
	FormatLPCM Format = "lpcm"
	FormatOGG  Format = "oggopus"
)

type YDX struct {
	FolderId        string
	OAuthToken      string
	IamToken        string
	IamTokenExpires time.Time
}

func New(OAuthToken, folderId string) *YDX {

	return &YDX{
		OAuthToken: OAuthToken,
		FolderId:   folderId,
	}
}

type IAMTokenRequest struct {
	YandexPassportOauthToken string `json:"yandexPassportOauthToken"`
}

type IAMTOkenRespons struct {
	IamToken  string `json:"iamToken"`
	ExpiresAt string `json:"expiresAt"`
}

func (y *YDX) getIAMToken() (IAMTOkenRespons, error) {
	sBody := IAMTOkenRespons{}

	data := IAMTokenRequest{
		YandexPassportOauthToken: y.OAuthToken,
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		return sBody, err
	}

	req, err := http.NewRequest(http.MethodPost, URIOAuthToken, bytes.NewBuffer(dataByte))
	if err != nil {
		return sBody, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return sBody, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return sBody, err
	}

	err = json.Unmarshal(body, &sBody)

	return sBody, err

}

func (y *YDX) UpdIamToken() error {
	fmt.Println("time duration", time.Until(y.IamTokenExpires), time.Hour)
	if y.IamToken != "" && time.Until(y.IamTokenExpires) > time.Hour {
		return nil
	}
	res, err := y.getIAMToken()
	if err != nil {
		return err
	}

	y.IamToken = res.IamToken
	y.IamTokenExpires, err = time.Parse("2006-01-02T15:04:05.999999999Z", res.ExpiresAt)
	if err != nil {
		return err
	}
	return nil
}

func (y *YDX) postRequest(url string, headers map[string]string, body interface{}) ([]byte, error) {
	var result []byte

	dataByte, err := json.Marshal(body)
	if err != nil {
		return result, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(dataByte))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+y.IamToken)
	for _, k := range headers {
		req.Header.Set(k, headers[k])
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	result, err = io.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	return result, err
}
