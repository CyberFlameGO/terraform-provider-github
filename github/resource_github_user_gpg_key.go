package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v52/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubUserGpgKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubUserGpgKeyCreate,
		Read:   resourceGithubUserGpgKeyRead,
		Delete: resourceGithubUserGpgKeyDelete,

		Schema: map[string]*schema.Schema{
			"armored_public_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Your public GPG key, generated in ASCII-armored format.",
			},
			"key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The key ID of the GPG key.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubUserGpgKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	pubKey := d.Get("armored_public_key").(string)
	ctx := context.Background()

	key, _, err := client.Users.CreateGPGKey(ctx, pubKey)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(key.GetID(), 10))

	return resourceGithubUserGpgKeyRead(d, meta)
}

func resourceGithubUserGpgKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	key, _, err := client.Users.GetGPGKey(ctx, id)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing user GPG key %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("key_id", key.GetKeyID())

	return nil
}

func resourceGithubUserGpgKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Users.DeleteGPGKey(ctx, id)

	return err
}
