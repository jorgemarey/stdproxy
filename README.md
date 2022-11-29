# Stdproxy

Stdproxy is a program that can be used to proxy stdin & stdout throught a configured socks proxy

## Usage

```text
./stdproxy -h
Usage of ./stdproxy:
  -basic-auth
        Use basic authentication (default is to read 'PROXY_CREDS' environment variable)
  -creds-file string
        Filepath of proxy credentials
  -log
        Enable logging
  -log-file string
        Save log execution to file (default "/tmp/stdproxy_1669706846.log")
  -timeout duration
        Proxy connection timeout (default 5s)
  -version
        Show program version
```

You can configure this on your ssh config in the ProxyCommand option

```text
Host foo_bar.com
  ProxyCommand /usr/local/bin/stdproxy proxy.internal.domain:3128 %h %p

Host foo_baz.es
  ProxyCommand /usr/local/bin/stdproxy --creds-file /tmp/secret.creds proxy.internal.domain:3128 %h %p 
```