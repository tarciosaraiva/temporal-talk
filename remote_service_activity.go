package temporaltalk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPGetter interface {
	Get(url string) (*http.Response, error)
}

type RemoteServiceActivity struct {
	HTTPClient HTTPGetter
}

func (i *RemoteServiceActivity) GetIP(ctx context.Context) (string, error) {
	resp, err := i.HTTPClient.Get("https://icanhazip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(body))
	return ip, nil
}

func (i *RemoteServiceActivity) GetLocationInfo(ctx context.Context, ip string) (*Geopoint, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := i.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := &Geopoint{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (i *RemoteServiceActivity) GetWeather(ctx context.Context, lat float32, lon float32) (float32, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.3f&longitude=%.3f&current=temperature", lat, lon)
	resp, err := i.HTTPClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	type currentWeather struct {
		Temperature float32 `json:"temperature"`
	}

	var data struct {
		Current currentWeather `json:"current"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}

	return data.Current.Temperature, nil
}
