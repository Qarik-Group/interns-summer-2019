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

	"github.com/coreos/etcd/pkg/transport"
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
			"endpoints" : "https://192.168.42.45:2379"   # REQUIRED

			"auth"  : false												# is role based or cert based auth enabled on the etcd cluster
			"user"  : "admin",            								# username for role based authentication
			"pass"  : "p@ssw0rd"          								# password for role based authentication
			"clientCertPath" : "/tmp/test-certs/test-client.crt" 		# path to client certificate
			"clientKeyPath"  : "/tmp/test-certs/test-client-key.key" 	# path to client key
			"caCertPath"     : "/tmp/test-certs/test-ca.crt" 			# path to CA certificate
			"fullOverwrite" 	 : "false"									# enable or disable full overwrite of the cluster
			"prefix"		 : "starkandwayne/"							# backup specific keys
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
				Title:    "Endpoint",
				Help:     "IP addresses of the etcd nodes in the cluster",
				Example:  "https://192.168.42.45:2379",
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
				Type:    "password",
				Help:    "Password for role based authentication",
				Title:   "Password",
				Example: "p@ssw0rd",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "clientCertPath",
				Type:    "string",
				Help:    "Path to the certificate issued by the CA for the client connecting to the ETCD cluster",
				Title:   "Client Certificate File Path",
				Example: "/tmp/test-certs/test-client.crt",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "clientKeyPath",
				Type:    "string",
				Help:    "Path to the key issued by the CA for the client connecting to the ETCD cluster",
				Title:   "Client Key File Path",
				Example: "/tmp/test-certs/test-client-key.key",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "caCertPath",
				Type:    "string",
				Help:    "Path to the CA certificate that issued the client cert and key",
				Title:   "Trusted CA File Path",
				Example: "/tmp/test-certs/test-ca.crt",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "fullOverwrite",
				Type:    "bool",
				Help:    "If this is enabled, the existing key/value pairs will be deleted. The values will be restored using the backup archive.",
				Title:   "Full Overwrite",
				Example: "false",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "prefix",
				Type:    "string",
				Help:    "This is the input string for prefix based backup-restore",
				Title:   "Prefix",
				Example: "starkandwayne/",
			},
		},
	}
	fmt.Fprintf(os.Stderr, "etcd plugin starting up...\n")
	plugin.Run(p)
}

type EtcdPlugin plugin.PluginInfo

type EtcdConfig struct {
	EtcdEndpoints  string
	Authentication bool
	Username       string
	Password       string
	ClientCertPath string
	ClientKeyPath  string
	CaCertPath     string
	FullOverwrite  bool
	Prefix         string
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

	clientCert, err := endpoint.StringValueDefault("clientCertPath", "")
	if err != nil {
		return nil, err
	}

	clientKey, err := endpoint.StringValueDefault("clientKeyPath", "")
	if err != nil {
		return nil, err
	}
	caCert, err := endpoint.StringValueDefault("caCertPath", "")
	if err != nil {
		return nil, err
	}

	fullOverwrite, err := endpoint.BooleanValueDefault("fullOverwrite", false)
	if err != nil {
		return nil, err
	}

	prefix, err := endpoint.StringValueDefault("prefix", "")
	if err != nil {
		return nil, err
	}

	return &EtcdConfig{
		EtcdEndpoints:  etcdEndpoint,
		Authentication: auth,
		Username:       user,
		Password:       pass,
		ClientCertPath: clientCert,
		ClientKeyPath:  clientKey,
		CaCertPath:     caCert,
		FullOverwrite:  fullOverwrite,
		Prefix:         prefix,
	}, nil
}

func (p EtcdPlugin) Validate(endpoint plugin.ShieldEndpoint) error {
	var (
		s    string
		b    bool
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

	b, err = endpoint.BooleanValueDefault("auth", false)
	if err != nil {
		fmt.Printf("@R{\u2717 authentication %s}\n", err)
		fail = true
	} else if b {
		fmt.Printf("@G{\u2713 auth} authentication is enabled\n")
	} else {
		fmt.Printf("@G{\u2713 auth} authentication is disabled\n")
	}

	if b {
		s, err = endpoint.StringValueDefault("user", "")
		if err != nil {
			fmt.Printf("@R{\u2717 user  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@G{\u2713 user} username was not provided so cert based auth will be used\n")
		} else {
			fmt.Printf("@G{\u2713 username} @C{%s}\n", plugin.Redact(s))
		}

		s, err = endpoint.StringValueDefault("pass", "")
		if err != nil {
			fmt.Printf("@R{\u2717 pass  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@R{\u2713 pass} password was not provided so cert based auth will be used\n")
		} else {
			fmt.Printf("@G{\u2713 password} @C{%s}\n", plugin.Redact(s))
		}

		s, err = endpoint.StringValueDefault("clientCertPath", "")
		if err != nil {
			fmt.Printf("@R{\u2717 client certificate path  %s}\n", err)
		} else if s == "" {
			fmt.Printf("@R{\u2713 client certificate path} was not provided\n")
		} else {
			fmt.Printf("@G{\u2713 client certificate path} was provided\n")
		}

		s, err = endpoint.StringValueDefault("clientKeyPath", "")
		if err != nil {
			fmt.Printf("@R{\u2717 client key path  %s}\n", err)
		} else if s == "" {
			fmt.Printf("@R{\u2713 client key path} was not provided\n")
		} else {
			fmt.Printf("@G{\u2713 client key path} was provided\n")
		}

		s, err = endpoint.StringValueDefault("caCertPath", "")
		if err != nil {
			fmt.Printf("@R{\u2717 CA certificate path  %s}\n", err)
		} else if s == "" {
			fmt.Printf("@R{\u2713 CA certificate path} was not provided\n")
		} else {
			fmt.Printf("@G{\u2713 CA certificate path} was provided\n")
		}
	}

	b, err = endpoint.BooleanValueDefault("fullOverwrite", false)
	if err != nil {
		fmt.Printf("@R{\u2717 full restore  %s}\n", err)
		fail = true
	} else if b {
		fmt.Printf("@G{\u2713} full restore enabled\n")
	} else {
		fmt.Printf("@G{\u2713} full restore disabled\n")
	}

	s, err = endpoint.StringValueDefault("prefix", "")
	if err != nil {
		fmt.Printf("@R{\u2717 prefix  %s}\n", err)
		fail = true
	} else if s == "" {
		fmt.Printf("@R{\u2713 prefix} prefix was not provided so everything will be backed-up/restored\n")
	} else {
		fmt.Printf("@G{\u2713 prefix} @C{%s}\n", plugin.Redact(s))
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

	tlsInfo := transport.TLSInfo{
		CertFile:      etcd.ClientCertPath,
		KeyFile:       etcd.ClientKeyPath,
		TrustedCAFile: etcd.CaCertPath,
	}

	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd.EtcdEndpoints},
		DialTimeout: 2 * time.Second,
		Username:    etcd.Username,
		Password:    etcd.Password,
		TLS:         tlsConfig,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer cli.Close()
	defer cancel()

	resp, err := cli.Get(ctx, etcd.Prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))

	if err != nil {
		log.Fatal(err)
	}

	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", base64.StdEncoding.EncodeToString([]byte(ev.Key)), base64.StdEncoding.EncodeToString([]byte(ev.Value)))
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

	tlsInfo := transport.TLSInfo{
		CertFile:      etcd.ClientCertPath,
		KeyFile:       etcd.ClientKeyPath,
		TrustedCAFile: etcd.CaCertPath,
	}

	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd.EtcdEndpoints},
		DialTimeout: 2 * time.Second,
		Username:    etcd.Username,
		Password:    etcd.Password,
		TLS:         tlsConfig,
	})

	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	if etcd.FullOverwrite {
		if etcd.Prefix == "" {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err := cli.Delete(ctx, "", clientv3.WithPrefix())
			if err != nil {
				log.Fatal(err)
			}
			defer cancel()
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err := cli.Delete(ctx, etcd.Prefix, clientv3.WithPrefix())
			if err != nil {
				log.Fatal(err)
			}
			defer cancel()
		}
	}

	for {
		line, buffer, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		lineBuffer := []byte{}
		lineBuffer = append(lineBuffer, line...)
		for buffer {
			partLine := []byte{}
			partLine, buffer, err = reader.ReadLine()
			if err == io.EOF {
				break
			}
			lineBuffer = append(lineBuffer, partLine...)
		}

		a := strings.Split(string(lineBuffer), " : ")
		strkey := a[0]
		strval := a[1]

		datakey, err := base64.StdEncoding.DecodeString(strkey)
		if err != nil {
			fmt.Printf("error decoding key: %s", err)
			return err
		}

		dataval, err := base64.StdEncoding.DecodeString(strval)
		if err != nil {
			fmt.Printf("error decoding value: %s", err)
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = cli.Put(ctx, fmt.Sprintf("%s", datakey), fmt.Sprintf("%s", dataval))
		defer cancel()
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
