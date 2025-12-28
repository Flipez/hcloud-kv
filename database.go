package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/klauspost/compress/zstd"
)

type Database struct {
	Store           map[string]string
	Name            string
	Client          *hcloud.Client
	Context         context.Context
	Self            *hcloud.Firewall
	NoInfo          bool
	LastEncodedSize int
}

func (d *Database) Init() {
	_, response, err := d.Client.Firewall.Create(d.Context, hcloud.FirewallCreateOpts{Name: d.Name})

	if err != nil {
		if response != nil && response.StatusCode == 409 {
			log.Printf("database %s already exists", d.Name)
		} else {
			log.Fatalf("unhandled error: %s\n", err)
		}
	} else {
		log.Printf("created new database: %s", d.Name)
	}

	d.Fetch()
}

func (d *Database) Fetch() *hcloud.Firewall {
	db, _, err := d.Client.Firewall.Get(d.Context, d.Name)
	if err != nil {
		log.Fatalf("error retrieving database: %s\n", err)
	}
	d.Self = db

	if d.Self != nil {
		store, size, err := rulesToMap(d.Self.Rules)
		if err != nil {
			log.Printf("could not parse rules (db might be empty or old format): %s", err)
			if d.Store == nil {
				d.Store = make(map[string]string)
			}
		} else {
			d.Store = store
			d.LastEncodedSize = size
		}
	}

	checkSize(d, d.LastEncodedSize)
	return d.Self
}

func (d *Database) Set(key, value string) bool {
	d.Fetch()

	if d.Store == nil {
		d.Store = make(map[string]string)
	}

	d.Store[key] = value

	firewallRules, encodedLength, err := mapToRules(d.Store)

	_, _, err = d.Client.Firewall.SetRules(d.Context, d.Self, hcloud.FirewallSetRulesOpts{Rules: firewallRules})

	if err != nil {
		log.Fatalf("error updating db: %s\n", err)
	}

	d.LastEncodedSize = encodedLength
	log.Println("OK")
	checkSize(d, encodedLength)

	return true
}

func (d *Database) Get(key string) string {
	d.Fetch()
	return d.Store[key]
}

func (d *Database) List() []string {
	d.Fetch()

	keys := make([]string, 0, len(d.Store))

	for k := range d.Store {
		keys = append(keys, k)
	}

	return keys
}

func checkSize(db *Database, currentLength int) {
	if db.NoInfo {
		return
	}

	maxChars := 500 * 255
	usage := (float64(currentLength) / float64(maxChars)) * 100

	log.Printf("[Info] Storage: %d/%d chars (%.2f%% used)", currentLength, maxChars, usage)
}

func rulesToMap(rules []hcloud.FirewallRule) (map[string]string, int, error) {
	store := map[string]string{}
	var sb strings.Builder
	sb.Grow(len(rules) * 255)

	for _, rule := range rules {
		if rule.Description != nil {
			sb.WriteString(*rule.Description)
		}
	}
	tempString := sb.String()
	totalLength := len(tempString)
	if totalLength == 0 {
		return store, 0, nil
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(tempString)
	if err != nil {
		return store, totalLength, err
	}

	zr, err := zstd.NewReader(bytes.NewReader(decodedBytes))
	if err != nil {
		return store, totalLength, err
	}
	defer zr.Close()

	err = gob.NewDecoder(zr).Decode(&store)
	return store, totalLength, err
}

func mapToRules(store map[string]string) ([]hcloud.FirewallRule, int, error) {
	var gobBuf bytes.Buffer
	if err := gob.NewEncoder(&gobBuf).Encode(store); err != nil {
		return nil, 0, err
	}

	var compressedBuf bytes.Buffer
	zw, _ := zstd.NewWriter(&compressedBuf)
	zw.Write(gobBuf.Bytes())
	zw.Close()

	encodedString := base64.StdEncoding.EncodeToString(compressedBuf.Bytes())
	totalLength := len(encodedString)

	var rules []hcloud.FirewallRule
	limit := 255
	_, dummyNet, _ := net.ParseCIDR("0.0.0.0/32")

	for i := 0; i < len(encodedString); i += limit {
		end := i + limit
		if end > len(encodedString) {
			end = len(encodedString)
		}

		chunk := encodedString[i:end] // Direct slice is safer for ASCII Base64

		rules = append(rules, hcloud.FirewallRule{
			Description: hcloud.Ptr(chunk),
			Direction:   hcloud.FirewallRuleDirectionIn,
			Protocol:    hcloud.FirewallRuleProtocolTCP,
			Port:        hcloud.Ptr("80"),
			SourceIPs:   []net.IPNet{*dummyNet},
		})
	}

	for i, r := range rules {
		fmt.Printf("Rule %d: Desc Length: %d, Protocol: %s\n", i, len(*r.Description), r.Protocol)
	}

	return rules, totalLength, nil
}
