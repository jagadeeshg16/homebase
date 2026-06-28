package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Cloudflare struct {
	APIToken string
	ZoneID   string
}

type cfRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type cfListResponse struct {
	Result []cfRecord `json:"result"`
}

type cfSingleResponse struct {
	Result cfRecord `json:"result"`
}

func (c *Cloudflare) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Cloudflare) listRecords(name string) ([]cfRecord, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", c.ZoneID, name)
	req, _ := http.NewRequest("GET", url, nil)
	c.auth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result cfListResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result, nil
}

func (c *Cloudflare) GetCurrentIP(domain, name string) (string, error) {
	fullName := name + "." + domain
	if name == "@" {
		fullName = domain
	}
	records, err := c.listRecords(fullName)
	if err != nil || len(records) == 0 {
		return "", err
	}
	return records[0].Content, nil
}

func (c *Cloudflare) UpsertARecord(domain, name, ip string, ttl int) error {
	fullName := name + "." + domain
	if name == "@" {
		fullName = domain
	}
	records, err := c.listRecords(fullName)
	if err != nil {
		return err
	}

	record := cfRecord{Type: "A", Name: fullName, Content: ip, TTL: ttl}
	var (
		method string
		url    string
		body   []byte
	)

	if len(records) > 0 {
		method = "PUT"
		url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.ZoneID, records[0].ID)
	} else {
		method = "POST"
		url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", c.ZoneID)
	}

	body, _ = json.Marshal(record)
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	c.auth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare error %d: %s", resp.StatusCode, b)
	}
	return nil
}

func (c *Cloudflare) DeleteRecord(domain, name string) error {
	fullName := name + "." + domain
	if name == "@" {
		fullName = domain
	}
	records, err := c.listRecords(fullName)
	if err != nil || len(records) == 0 {
		return err
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.ZoneID, records[0].ID)
	req, _ := http.NewRequest("DELETE", url, nil)
	c.auth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare delete error %d: %s", resp.StatusCode, b)
	}
	return nil
}
