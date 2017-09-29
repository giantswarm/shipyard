package shipyard

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	kubeconfig = `
apiVersion: v1
clusters:
- cluster:
    server: http://127.0.0.1:8080
  name: local
contexts:
- context:
    cluster: local
  name: local
current-context: local
kind: Config
preferences: {}
`
	caCrt = `
-----BEGIN CERTIFICATE-----
MIIDUDCCAjigAwIBAgIJAJcxePJjlT7KMA0GCSqGSIb3DQEBCwUAMB8xHTAbBgNV
BAMMFDEyNy4wLjAuMUAxNTA2NTkzMDM4MB4XDTE3MDkyODEwMDM1OVoXDTI3MDky
NjEwMDM1OVowHzEdMBsGA1UEAwwUMTI3LjAuMC4xQDE1MDY1OTMwMzgwggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDc3oP6y6OlgxfnNpm8kAiYXUFzvYyj
4JSY84FbACrNfV05TDLqWN5LY4doRfMldWCb/91EPQ1y9ChM8tcSl1C9gV8eOKBT
0caELrAvmU5SANeOxeSmXPEJNN8CUcrpm46gatCgOZVK20rhRHSZUwxOUWaXxQ/q
dp4Fxg0UwdVaiAOGItn9/IO/5Z1Cnp/eDO3vIq9HW/W9OmOwSy1fkH5WyozUEjpK
R5lrQ23Ex00CfZes41TBpMtEgI2yZ/9q8lziJ8I4GOzwOlZhpsyURWxV/MtpU+jW
oqaJZkO3p1LvJKR3800htc3u3PaGMOrwdUDvLKhj+AaRmFhGagU3jgRFAgMBAAGj
gY4wgYswHQYDVR0OBBYEFN0AiVRktY+9Z7Xwj6tj11vdCaGhME8GA1UdIwRIMEaA
FN0AiVRktY+9Z7Xwj6tj11vdCaGhoSOkITAfMR0wGwYDVQQDDBQxMjcuMC4wLjFA
MTUwNjU5MzAzOIIJAJcxePJjlT7KMAwGA1UdEwQFMAMBAf8wCwYDVR0PBAQDAgEG
MA0GCSqGSIb3DQEBCwUAA4IBAQCe/QiHxV+ESDLFJu2KoR47tHwMks407uRiTSM8
O1oT5dnBzz/2QPdvqTNTk+km8r0mOjoNf7betF3pPV3dQkyf29ItjuYOr9bOgrxr
fx0FmDl6Re2bvEPgyEh89cvAnvRTAp2rxowxLJQehbIsU18cnML6WZ0K0XWaza97
Me25k77S0iT5L034r+AX9dbSyLn55yExoKc89nvD10llGxFNJ1WxNPlZmtU8MSQR
KpA8U4yEnxP5d4uWPck7jgdVj2WTrttYQRDqE5MuIR9SaG3URymuS7/N4HXIXO6p
6rUfPfVamFvxdg/T36kG3jTSOgANMz5RfyMvSJek6VU2t9Ac
-----END CERTIFICATE-----
`
	serverCert = `
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 1 (0x1)
    Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN=127.0.0.1@1506593038
        Validity
            Not Before: Sep 28 10:03:59 2017 GMT
            Not After : Sep 26 10:03:59 2027 GMT
        Subject: CN=kubernetes-master
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:bc:8e:3d:a0:ad:d8:ea:e0:9c:f6:74:7e:f9:e7:
                    57:e2:cb:c0:8b:6f:1c:8d:33:a5:4c:e0:95:62:b0:
                    e9:17:8d:c1:08:58:38:e9:01:f4:04:1c:f4:ff:02:
                    53:31:5c:3a:8a:fa:56:2c:05:95:0a:4a:8a:e2:bf:
                    9d:18:ed:bd:64:16:29:6e:26:1b:63:88:91:b2:18:
                    8b:8b:cd:d6:c3:83:f8:f1:09:a3:e0:dd:66:24:3b:
                    ac:e6:5e:79:59:ad:be:be:42:3c:70:f1:a7:62:57:
                    5b:20:de:79:d8:86:fb:90:23:21:72:d2:3f:a5:fe:
                    cc:f6:47:a3:2d:81:c1:75:cc:e9:98:26:3a:60:af:
                    4e:c9:3f:8e:75:6e:e4:1a:3e:44:76:60:5e:07:9c:
                    9b:38:35:58:ff:99:c4:16:9a:29:0b:29:cf:33:f9:
                    a6:7c:94:2e:88:f4:d8:18:71:6f:2f:1c:a4:76:2d:
                    01:90:42:ed:50:6b:7d:8f:1d:f0:af:63:c3:da:79:
                    ee:88:2d:8a:fa:0b:a2:42:f3:fc:5c:60:27:c5:34:
                    8e:b3:c1:f6:81:bf:d3:07:83:4e:49:5b:8a:bc:04:
                    25:e6:43:d5:93:0e:38:37:d3:4d:e9:b9:f6:eb:6e:
                    ae:c2:02:1b:03:41:2b:e7:26:6c:2d:00:60:05:f5:
                    86:7b
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Basic Constraints:
                CA:FALSE
            X509v3 Subject Key Identifier:
                34:4D:54:A5:14:2D:88:7D:A0:87:94:50:BE:1E:91:7C:8F:9E:8B:22
            X509v3 Authority Key Identifier:
                keyid:DD:00:89:54:64:B5:8F:BD:67:B5:F0:8F:AB:63:D7:5B:DD:09:A1:A1
                DirName:/CN=127.0.0.1@1506593038
                serial:97:31:78:F2:63:95:3E:CA

            X509v3 Extended Key Usage:
                TLS Web Server Authentication
            X509v3 Key Usage:
                Digital Signature, Key Encipherment
            X509v3 Subject Alternative Name:
                IP Address:127.0.0.1
    Signature Algorithm: sha256WithRSAEncryption
         13:40:3e:48:46:27:7e:c4:53:f0:37:e1:1a:86:89:1c:ca:c3:
         c8:b4:e3:6b:0e:3c:71:57:a9:e3:30:9c:90:14:cd:8f:b9:53:
         5f:69:d1:7c:43:1c:2d:b0:38:b9:23:1b:5a:f5:9a:bb:6e:3e:
         9d:9a:ae:2d:02:a4:19:29:10:1a:88:a7:96:c1:77:e0:77:17:
         93:f2:27:88:e1:06:44:3f:49:fa:10:d9:6d:5c:60:31:84:55:
         0e:fc:c9:b2:f9:ac:f9:b2:a4:84:61:92:e4:86:e3:64:96:c5:
         cc:6a:c3:8c:60:71:33:40:c1:d3:92:22:c1:e7:24:63:9e:d9:
         db:2d:66:d0:a6:ed:d7:d7:45:eb:a5:c9:14:b3:be:6a:55:32:
         64:e4:76:57:84:98:fb:eb:98:06:ae:27:ea:b6:e2:fc:08:8c:
         a3:3a:fe:30:6f:3a:9b:25:67:9f:b2:35:c5:80:ee:4b:1b:98:
         81:9e:b5:be:f1:19:8b:a9:96:ed:f6:e0:7e:60:5f:ee:a9:31:
         3a:1a:a7:43:db:4d:3a:4f:73:2b:08:9e:fc:42:2b:db:0e:51:
         40:61:d8:ae:8b:05:67:7e:55:23:50:ab:35:59:cb:2b:7f:7f:
         b0:ff:d9:0c:89:c5:51:a9:51:eb:a0:2c:c9:dc:f4:8e:4c:ff:
         33:93:de:ba
-----BEGIN CERTIFICATE-----
MIIDaDCCAlCgAwIBAgIBATANBgkqhkiG9w0BAQsFADAfMR0wGwYDVQQDDBQxMjcu
MC4wLjFAMTUwNjU5MzAzODAeFw0xNzA5MjgxMDAzNTlaFw0yNzA5MjYxMDAzNTla
MBwxGjAYBgNVBAMMEWt1YmVybmV0ZXMtbWFzdGVyMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAvI49oK3Y6uCc9nR++edX4svAi28cjTOlTOCVYrDpF43B
CFg46QH0BBz0/wJTMVw6ivpWLAWVCkqK4r+dGO29ZBYpbiYbY4iRshiLi83Ww4P4
8Qmj4N1mJDus5l55Wa2+vkI8cPGnYldbIN552Ib7kCMhctI/pf7M9kejLYHBdczp
mCY6YK9OyT+OdW7kGj5EdmBeB5ybODVY/5nEFpopCynPM/mmfJQuiPTYGHFvLxyk
di0BkELtUGt9jx3wr2PD2nnuiC2K+guiQvP8XGAnxTSOs8H2gb/TB4NOSVuKvAQl
5kPVkw44N9NN6bn2626uwgIbA0Er5yZsLQBgBfWGewIDAQABo4GxMIGuMAkGA1Ud
EwQCMAAwHQYDVR0OBBYEFDRNVKUULYh9oIeUUL4ekXyPnosiME8GA1UdIwRIMEaA
FN0AiVRktY+9Z7Xwj6tj11vdCaGhoSOkITAfMR0wGwYDVQQDDBQxMjcuMC4wLjFA
MTUwNjU5MzAzOIIJAJcxePJjlT7KMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAsGA1Ud
DwQEAwIFoDAPBgNVHREECDAGhwR/AAABMA0GCSqGSIb3DQEBCwUAA4IBAQATQD5I
Rid+xFPwN+EahokcysPItONrDjxxV6njMJyQFM2PuVNfadF8QxwtsDi5Ixta9Zq7
bj6dmq4tAqQZKRAaiKeWwXfgdxeT8ieI4QZEP0n6ENltXGAxhFUO/Mmy+az5sqSE
YZLkhuNklsXMasOMYHEzQMHTkiLB5yRjntnbLWbQpu3X10XrpckUs75qVTJk5HZX
hJj765gGrifqtuL8CIyjOv4wbzqbJWefsjXFgO5LG5iBnrW+8RmLqZbt9uB+YF/u
qTE6GqdD2006T3MrCJ78QivbDlFAYdiuiwVnflUjUKs1Wcsrf3+w/9kMicVRqVHr
oCzJ3PSOTP8zk966
-----END CERTIFICATE-----
`
	serverKey = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC8jj2grdjq4Jz2
dH7551fiy8CLbxyNM6VM4JVisOkXjcEIWDjpAfQEHPT/AlMxXDqK+lYsBZUKSori
v50Y7b1kFiluJhtjiJGyGIuLzdbDg/jxCaPg3WYkO6zmXnlZrb6+Qjxw8adiV1sg
3nnYhvuQIyFy0j+l/sz2R6MtgcF1zOmYJjpgr07JP451buQaPkR2YF4HnJs4NVj/
mcQWmikLKc8z+aZ8lC6I9NgYcW8vHKR2LQGQQu1Qa32PHfCvY8Paee6ILYr6C6JC
8/xcYCfFNI6zwfaBv9MHg05JW4q8BCXmQ9WTDjg3003pufbrbq7CAhsDQSvnJmwt
AGAF9YZ7AgMBAAECggEAWrkh4+bh4RlTzK1+zuVF/yTELxK2LSZ1WYkRt5uhe6jE
ASzZrRs4eqYoJ27J8o0KygXrYmEJNhtpufIyN2VkY/zZ0FrbgMecOpHeytSuigI8
zFU0GaTNWY+xAGLihoi3pzmddUaAoCuh+C2zeMBx2AdfN6z73PW9Tp5nYCT3naK4
MV1oTR7YPtqtvzHRw1NyHIyMM7fHQwxjAa6hKdvdIv0IjLf+cK0Nao8kkw56QbYq
jye4TQNGZzs4Usc78Ehqj2zQLE9pxLBLOpv03CgQtpJe5BeiL0nDowzZIc+oiTrB
TJ9+8rgWLny6Kd0PS6C9x4OPguclN+HuRtOe/LvNwQKBgQDclqxmZ43ED8JgqBtE
OvrowbU7RY+pySrghlTGRF5MAvcHGpj7to7Ku/C2432nLqrp+WM3v1XS80MvNwle
l+JhmezDEPL1jQ2BKf0ubw+qV9kZo46ogbrFUHWMnWbpo70WvwMxi3jzogvsG9Lq
HSmSbEgNanrw5y6H7HZ/YacN2wKBgQDa0yOMK0cWPjoK1jhMTKAdzWiEfoVt+xI/
ZnZBc8qWGA1obMQmTpYLTfDvwW0HOAMFzt81lT4ZAQYdV0e8UegoKmzxVAv1aOJd
YMVA5mXoVZHGNjsxj8pRJg3DXwjhzuJYXNo0i3/9fKUPWSYc8qP0oBTG46EnAw9t
AHGrcljb4QKBgC7bLpm+C2YHNvHTI5+Vq7B/XSDPANo+6gWxYxaOdT1OL+zpYG+v
cptr2pDut8Uoa5OxrrqrwO5DUBUaaroWJzc2PA2fbwxrvt+d7LLNUpWLfYktreLr
U6IQGjgZQ0AD0Omg/2upxbJyzHeF3YJvWWJJ7/AxmxXK9Z5Xw0ABnTubAoGBAIsj
K3wp3HZ5NKDFW2CwbDLm8+kjJaYruYuUk+bEQHE1c/kNB5+v4lnnwiZAoBmx9MIR
qv3AGo79hqzLXXKRxgMcDs9X+I6flSd4q5O7q9qR5jHZM8QswKDeiGvMlrI1wNgc
miZE+SntwmpC7igD5FpcGznnbQWIPZu6Z4xzFashAoGAGKY0b4rQmPU13Zw1AQPB
Z4QJb4zuF2UTgVd8jbwreIjDyOkdLwhMoy5RkzORNumcvjZa+WgZ+CVQ7F3PaEV1
Hzh21hrkQHbBr7cXFqnbAjvmqfUBpeXokdkJ39QzoXrKtiiN1cO7XHV3wnhaMmUi
mUI8p41+CZ+edrvtyS1Sj4s=
-----END PRIVATE KEY-----
`
	addonManagerManifest = `
{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "kube-addon-manager",
		"namespace": "kube-system",
		"version": "v1"
	},
	"spec": {
		"hostNetwork": true,
		"containers": [
			{
				"name": "kube-addon-manager",
				"image": "gcr.io/google_containers/kube-addon-manager-amd64:v6.1",
				"resources": {
					"requests": {
						"cpu": "5m",
						"memory": "50Mi"
					}
				},
				"volumeMounts": [
					{
						"name": "addons",
						"mountPath": "/etc/kubernetes/addons",
						"readOnly": false
					}
				]
			}
		],
		"volumes": [
			{
				"name": "addons",
				"emptyDir": {}
			}
		]
	}
}
`
	masterManifest = `
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "k8s-master",
    "namespace": "kube-system"
  },
  "spec": {
    "hostNetwork": true,
    "containers": [
      {
        "name": "controller-manager",
        "image": "gcr.io/google_containers/hyperkube-amd64:v1.7.6",
        "command": [
          "/hyperkube",
          "controller-manager",
          "--master=127.0.0.1:8080",
          "--service-account-private-key-file=/srv/kubernetes/server.key",
          "--root-ca-file=/srv/kubernetes/ca.crt",
          "--min-resync-period=3m",
          "--leader-elect=true",
          "--v=2"
        ],
        "volumeMounts": [
          {
            "name": "srvkube",
            "mountPath": "/srv/kubernetes"
          }
        ]
      },
      {
        "name": "scheduler",
        "image": "gcr.io/google_containers/hyperkube-amd64:v1.7.6",
        "command": [
          "/hyperkube",
          "scheduler",
          "--master=127.0.0.1:8080",
          "--leader-elect=true",
          "--v=2"
        ]
      }
    ],
    "volumes": [
      {
        "name": "srvkube",
          "hostPath": {
            "path": "/srv/kubernetes"
          }
      }
    ]
  }
}
`
)

// PrepareBaseDir creates the base directory and required files for shipyard to run
func PrepareBaseDir() (string, error) {
	baseDir, err := ioutil.TempDir("", "gs-shipyard")
	if err != nil {
		return "", err
	}

	files := map[string]string{
		"test/e2e/cluster/config":                       kubeconfig,
		"test/e2e/cluster/cert/ca.crt":                  caCrt,
		"test/e2e/cluster/cert/server.cert":             serverCert,
		"test/e2e/cluster/cert/server.key":              serverKey,
		"test/e2e/cluster/manifests/addon-manager.json": addonManagerManifest,
		"test/e2e/cluster/manifests/master.json":        masterManifest,
	}

	for relPath, content := range files {
		fullPath := filepath.Join(baseDir, relPath)
		if err := os.MkdirAll(path.Dir(fullPath), os.ModePerm); err != nil {
			return "", err
		}

		if err := ioutil.WriteFile(fullPath, []byte(content), os.ModePerm); err != nil {
			return "", err
		}
	}

	return baseDir, nil
}
