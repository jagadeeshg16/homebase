package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GoDaddy struct {
	APIKey    string
	APISecret string
}

type gdRecord struct {
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
	Name string `json:"name"`
}

func (g *GoDaddy) auth() string {
	return fmt.Sprintf("sso-key %s:%s", g.APIKey, g.APISecret)
}

func (g *GoDaddy) GetCurrentIP(domain, name string) (string, error) {
	url := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/A/%s", domain, name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", g.auth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var records []gdRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", nil
	}
	return records[0].Data, nil
}

func (g *GoDaddy) UpsertARecord(domain, name, ip string, ttl int) error {
	url := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/A/%s", domain, name)
	body, _ := json.Marshal([]gdRecord{{Data: ip, TTL: ttl, Type: "A", Name: name}})

	req, _ := http.NewRequest("PUT", url, bytes.NewReader(body))
	req.Header.Set("Authorization", g.auth())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("godaddy error %d: %s", resp.StatusCode, b)
	}
	return nil
}

func (g *GoDaddy) DeleteRecord(domain, name string) error {
	url := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/A/%s", domain, name)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", g.auth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("godaddy delete error %d: %s", resp.StatusCode, b)
	}
	return nil
}
