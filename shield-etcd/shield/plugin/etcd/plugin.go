package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"log"
	"os"
	"strings"
	"time"

	fmt "github.com/jhunt/go-ansi"
	"go.etcd.io/etcd/clientv3"

	"github.com/starkandwayne/shield/plugin"
)

func main() {
	p := EtcdPlugin{
		Name:    "Etcd Backup Plugin",
		Author:  "Stark and Wayne",
		Version: "0.0.1",
		Features: plugin.PluginFeatures{
			Target: "yes",
			Store:  "no",
		},
		Example: `
{
  "endpoints" : "https://192.168.42.45:2379,https://192.168.42.46:2379,https://192.168.42.47:2379"   # REQUIRED

  "auth"  : false				# is role based or cert based auth enabled on the etcd cluster
  "user"  : "admin",            # username for role based authentication
  "pass"  : "p@ssw0rd"          # password for role based authentication
}
`,
		Defaults: `
{
}
`,
		Fields: []plugin.Field{
			plugin.Field{
				Mode:     "target",
				Name:     "endpoints",
				Type:     "string",
				Title:    "Endpoints",
				Help:     "IP addresses of the etcd nodes in the cluster",
				Example:  "https://192.168.42.45:2379,https://192.168.42.46:2379,https://192.168.42.47:2379",
				Required: true,
			},
			plugin.Field{
				Mode:    "target",
				Name:    "auth",
				Type:    "bool",
				Title:   "Authentication",
				Help:    "Is role based or cert based authentication enabled on the etcd cluster",
				Example: "false",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "user",
				Type:    "string",
				Help:    "Username for role based authentication",
				Title:   "Username",
				Example: "admin",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "pass",
				Type:    "string",
				Help:    "Password for role based authentication",
				Title:   "Password",
				Example: "p@ssw0rd",
			},
		},
	}
	fmt.Fprintf(os.Stderr, "etcd plugin starting up...\n")
	plugin.Run(p)
}

type EtcdPlugin plugin.PluginInfo

type EtcdConfig struct {
	etcdEndpoints  string
	Authentication bool
	Username       string
	Password       string
}

func (p EtcdPlugin) Meta() plugin.PluginInfo {
	return plugin.PluginInfo(p)
}

func getEtcdConfig(endpoint plugin.ShieldEndpoint) (*EtcdConfig, error) {
	etcdEndpoint, err := endpoint.StringValue("endpoints")
	if err != nil {
		return nil, err
	}

	auth, err := endpoint.BooleanValueDefault("auth", false)
	if err != nil {
		return nil, err
	}

	user, err := endpoint.StringValueDefault("user", "")
	if err != nil {
		return nil, err
	}

	pass, err := endpoint.StringValueDefault("pass", "")
	if err != nil {
		return nil, err
	}

	return &EtcdConfig{
		etcdEndpoints:  etcdEndpoint,
		Authentication: auth,
		Username:       user,
		Password:       pass,
	}, nil
}

func (p EtcdPlugin) Validate(endpoint plugin.ShieldEndpoint) error {
	var (
		s    string
		auth bool
		err  error
		fail bool
	)
	s, err = endpoint.StringValue("endpoints")
	if err != nil {
		fmt.Printf("@R{\u2717 endpoints  %s}\n", err)
		fail = true
	} else {
		fmt.Printf("@G{\u2713 etcd endpoints} data in @C{%s} will be backed up\n", s)
	}

	auth, err = endpoint.BooleanValueDefault("auth", false)
	if err != nil {
		fmt.Printf("@R{\u2717 authentication %s}\n", err)
		fail = true
	} else if auth {
		fmt.Printf("@G{\u2713 auth} authentication is enabled\n")
	} else {
		fmt.Printf("@G{\u2713 auth} authentication is disabled\n")
	}

	if auth {
		s, err = endpoint.StringValueDefault("user", "")
		if err != nil {
			fmt.Printf("@R{\u2717 user  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@G{\u2717 user} auth was enabled but username was not provided\n")
			fail = true
		} else {
			fmt.Printf("@G{\u2713 username} @C{%s}\n", plugin.Redact(s))
		}

		s, err = endpoint.StringValueDefault("pass", "")
		if err != nil {
			fmt.Printf("@R{\u2717 pass  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@R{\u2717 pass} auth was enabled but password was not provided\n")
			fail = true
		} else {
			fmt.Printf("@G{\u2713 password} @C{%s}\n", plugin.Redact(s))
		}
	}

	if fail {
		return fmt.Errorf("etcd: invalid configuration")
	}

	return nil
}

// Backup ETCD data
func (p EtcdPlugin) Backup(endpoint plugin.ShieldEndpoint) error {
	etcd, err := getEtcdConfig(endpoint)
	if err != nil {
		return err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd.etcdEndpoints},
		DialTimeout: 2 * time.Second,
		Username:    etcd.Username,
		Password:    etcd.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	resp, err := cli.Get(ctx, "", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	if err != nil {
		log.Fatal(err)
	}

	for _, ev := range resp.Kvs {
		dataK := []byte(ev.Key)
		dataV := []byte(ev.Value)
		encodedK := base64.StdEncoding.EncodeToString(dataK)
		encodedV := base64.StdEncoding.EncodeToString(dataV)
		newLine := fmt.Sprintf("%s : %s", encodedK, encodedV)
		cmd := fmt.Sprintf("echo %s", newLine)
		plugin.Exec(cmd, plugin.STDOUT)
	}
	return nil
}

// Restore ETCD data
func (p EtcdPlugin) Restore(endpoint plugin.ShieldEndpoint) error {
	reader := bufio.NewReader(os.Stdin)

	etcd, err := getEtcdConfig(endpoint)
	if err != nil {
		return err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd.etcdEndpoints},
		DialTimeout: 2 * time.Second,
		Username:    etcd.Username,
		Password:    etcd.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		a := strings.Split(string(line), " : ")
		strkey := a[0]
		strval := a[1]

		datakey, err := base64.StdEncoding.DecodeString(strkey)
		dataval, err := base64.StdEncoding.DecodeString(strval)
		if err != nil {
			fmt.Printf("error: %s", err)
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = cli.Put(ctx, fmt.Sprintf("%s", datakey), fmt.Sprintf("%s", dataval))
		cancel()
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (p EtcdPlugin) Store(endpoint plugin.ShieldEndpoint) (string, int64, error) {
	return "", 0, plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Retrieve(endpoint plugin.ShieldEndpoint, file string) error {
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Purge(endpoint plugin.ShieldEndpoint, key string) error {
	return plugin.UNIMPLEMENTED
}
