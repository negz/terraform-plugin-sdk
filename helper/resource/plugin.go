package resource

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	grpcplugin "github.com/hashicorp/terraform-plugin-sdk/v2/internal/helper/plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

func runProviderCommand(f func() error, wd *tftest.WorkingDir, opts *plugin.ServeOpts) error {
	// Run the provider in the same process as the test runner using the
	// reattach behavior in Terraform. This ensures we get test coverage
	// and enables the use of delve as a debugger.

	// the provider name is technically supposed to be specified in the
	// format returned by addrs.Provider.GetDisplay(), but 1. I'm not
	// importing the entire addrs package for this and 2. we only get the
	// provider name here. Fortunately, when only a provider name is
	// specified in a provider block--which is how the config file we
	// generate does things--Terraform just automatically assumes it's in
	// the hashicorp namespace and the default registry.terraform.io host,
	// so we can just construct the output of GetDisplay() ourselves, based
	// on the provider name. GetDisplay() omits the default host, so for
	// our purposes this will always be hashicorp/PROVIDER_NAME.
	providerName := wd.GetHelper().GetPluginName()

	// providerName gets returned as terraform-provider-foo, and we need
	// just foo. So let's fix that.
	providerName = strings.TrimPrefix(providerName, "terraform-provider-")

	// set up a context we can cancel, and defer the cancellation. This
	// will ensure the go-plugin Server doesn't block indefinitely; we're
	// going to use this context for it, and it knows to return when the
	// context is canceled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// set up a channel for go-plugin's ReattachConfig type. When the
	// server gets up and running, this is how we'll get its connection
	// info.
	reattachCh := make(chan *goplugin.ReattachConfig)

	// set up a close channel. When the go-plugin Server is done, it'll
	// close this for us, so we can block on it to make sure the go-plugin
	// Server returned.
	closeCh := make(chan struct{})

	// set up our ServeTestConfig. This tells go-plugin that we're going to
	// manage the lifecycle of the plugin, and we're going to take care of
	// orchestrating it. It prevents it from overwriting
	// os.Stdout/os.Stderr, prevents clients from killing the server, makes
	// the server a lot less noisy, and does a bunch of other stuff we
	// want.
	opts.TestConfig = &goplugin.ServeTestConfig{
		Context:          ctx,
		ReattachConfigCh: reattachCh,
		CloseCh:          closeCh,
	}

	// if we didn't override the logger, let's set a default one.
	if opts.Logger == nil {
		opts.Logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugintest",
			Level:  hclog.Trace,
			Output: ioutil.Discard,
		})
	}

	// this is needed so Terraform doesn't default to expecting protocol 4;
	// we're skipping the handshake because Terraform didn't launch the
	// plugin.
	os.Setenv("PLUGIN_PROTOCOL_VERSIONS", "5")

	// actually run the provider! Woo
	go plugin.Serve(opts)

	// ok, now we need to know how to connect to the provider. The provider
	// will tell us.
	var config *goplugin.ReattachConfig
	select {
	case config = <-reattachCh:
	case <-time.After(2 * time.Second):
		return errors.New("timeout waiting on reattach config")
	}

	if config == nil {
		return errors.New("nil reattach config received")
	}

	// when we tell Terraform how to connect, we do that with a
	// TF_REATTACH_PROVIDERS environment variable, the value of which is a
	// map of provider display names to reattach configs.
	type reattachConfig struct {
		Protocol string
		Addr     struct {
			Network string
			String  string
		}
		Pid  int
		Test bool
	}
	reattachStr, err := json.Marshal(map[string]reattachConfig{
		"hashicorp/" + providerName: reattachConfig{
			Protocol: string(config.Protocol),
			Addr: struct {
				Network string
				String  string
			}{
				Network: config.Addr.Network(),
				String:  config.Addr.String(),
			},
			Pid:  config.Pid,
			Test: config.Test,
		},
	})
	if err != nil {
		return err
	}
	wd.Setenv("TF_REATTACH_PROVIDERS", string(reattachStr))

	// ok, let's call whatever Terraform command the test was trying to
	// call, now that we know it'll attach back to that server we just
	// started.
	err = f()
	if err != nil {
		log.Printf("[WARN] Got error running Terraform: %s", err)
	}

	// cancel the server so it'll return. Otherwise, this closeCh won't get
	// closed, and we'll hang here.
	cancel()

	// wait for the server to actually shut down; it may take a moment for
	// it to clean up, or whatever.
	<-closeCh

	// once we've run the Terraform command, let's remove the reattach
	// information from the WorkingDir's environment. The WorkingDir will
	// persist until the next call, but the server in the reattach info
	// doesn't exist anymore at this point, so the reattach info is no
	// longer valid. In theory it should be overwritten in the next call,
	// but just to avoid any confusing bug reports, let's just unset the
	// environment variable altogether.
	wd.Unsetenv("TF_REATTACH_PROVIDERS")

	// return any error returned from the orchestration code running
	// Terraform commands
	return err
}

// defaultPluginServeOpts builds ths *plugin.ServeOpts that you usually want to
// use when running runProviderCommand. It just sets the ProviderFunc to return
// the provider under test.
func defaultPluginServeOpts(wd *tftest.WorkingDir, providers map[string]*schema.Provider) *plugin.ServeOpts {
	return &plugin.ServeOpts{
		ProviderFunc: acctest.TestProviderFunc,
		GRPCProviderFunc: func() proto.ProviderServer {
			return grpcplugin.NewGRPCProviderServer(acctest.TestProviderFunc())
		},
	}
}
