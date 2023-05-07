package main

import (
	"context"
	"log"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

type Database struct {
	Store   map[string]string
	Name    string
	Client  *hcloud.Client
	Context context.Context
	Self    *hcloud.SSHKey
}

func (d *Database) Init() {
	keyPlaceholder := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAUqWKFtn3P3p0tOWXMgkfc7aTc5Z17+LSlf50X/ep/Z"

	database, response, err := d.Client.SSHKey.Create(d.Context, hcloud.SSHKeyCreateOpts{Name: d.Name, PublicKey: keyPlaceholder})
	if err != nil && response.StatusCode == 409 {
		log.Fatalf("database %s already exists", d.Name)
	} else if err != nil {
		log.Fatalf("unhandled error: %s\n", err)
	}

	d.Store = database.Labels
}

func (d *Database) Fetch() *hcloud.SSHKey {
	db, _, err := d.Client.SSHKey.Get(d.Context, d.Name)
	d.Self = db

	if err != nil {
		log.Fatalf("error retrieving database: %s\n", err)
	}

	if d.Self != nil {
		d.Store = d.Self.Labels
	} else {
		log.Fatalf("database %s not found", d.Name)
	}

	return d.Self
}

func (d *Database) Set(key, value string) bool {
	d.Fetch()
	d.Store[key] = value

	database, _, err := d.Client.SSHKey.Update(d.Context, d.Self, hcloud.SSHKeyUpdateOpts{Labels: d.Store})

	if err != nil {
		log.Fatalf("error updating db: %s\n", err)
	}

	d.Store = database.Labels

	return true
}

func (d *Database) Get(key string) string {
	d.Fetch()
	return d.Store[key]
}

func (d *Database) List() []string {
	d.Fetch()

	keys := make([]string, len(d.Store))

	i := 0
	for k := range d.Store {
		keys[i] = k
		i++
	}

	return keys
}
