# proxy_checker

Take a list of addresses, check the proxy status, type, country, state, connection speed.

## Install 

```
go get github.com/13ph03nix/proxy_checker
```

## Usage

```
➜ proxy_checker -h
Usage of proxy_checker:
  -c int
    	set the concurrency level (default 128)
  -s	only socks5, skip the http proxy check
  -t int
    	timeout (milliseconds) (default 10000)


➜ curl "https://api.proxyscrape.com/?request=getproxies&proxytype=socks5&timeout=10000&country=all" | proxy_checker -s -t 5000
  148.*.*.93:9150    --  socks5                 --  Germany                --  Hesse                  --  2.322577837s
  47.*.*.169:24808   --  socks5                 --  United States          --  Minnesota              --  4.436653077s
  184.*.*.10:1158    --  socks5                 --  United States          --  Arizona                --  4.725385771s
  193.*.*.221:3327   --  socks5                 --  Russia                 --  Krasnodarskiy Kray     --  1.471982062s
  80.*.*.126:8975    --  socks5                 --  Seychelles             --                         --  4.698932901s
  222.*.*.245:3468   --  socks5                 --  Indonesia              --  Jakarta                --  3.743092558s
  216.*.*.21:5745    --  socks5                 --  United States          --  California             --  4.433900367s
```

## Credits

* [httprobe](https://github.com/tomnomnom/httprobe)
