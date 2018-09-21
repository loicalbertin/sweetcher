# Sweetcher: For those who know the hell of enterprise proxies


[![Build Status](https://travis-ci.org/loicalbertin/sweetcher.svg?branch=master)](https://travis-ci.org/loicalbertin/sweetcher) [![Go Report Card](https://goreportcard.com/badge/github.com/loicalbertin/sweetcher)](https://goreportcard.com/report/github.com/loicalbertin/sweetcher) [![GoDoc](https://godoc.org/github.com/loicalbertin/sweetcher?status.svg)](https://godoc.org/github.com/loicalbertin/sweetcher) [![codecov](https://codecov.io/gh/loicalbertin/sweetcher/branch/master/graph/badge.svg)](https://codecov.io/gh/loicalbertin/sweetcher) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Floicalbertin%2Fsweetcher.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Floicalbertin%2Fsweetcher?ref=badge_shield)

Sweetcher is a tool inspired from web browsers' proxy switchers plugins like SwitchyOmega or FoxyProxy but witch operate at OS level rather than only for your browser.

It allows a set of URL patterns to proxies defined in different profiles. And allows to easily switch from one profile to another.

## Here is a little use-case (true story!)

Lets say that you are working in a big IT company. This company does not allow its employees to access directly to the net, they have to go through a proxy system called masterproxy.yourcompany.it. This proxy has the one that all employees should use, when it is not out of order it has almost pretty good performances, but unfortunately it blocks some web sites like all file sharing files (google drive, dropbox, github gist (!?!?), ...) or some news sites like reedit (?!?!).
Fortunately you know another proxy (hiddenproxy.yourcompany.it) without blacklists but with poorest performances.

Sometimes you also do homeworking, in this case you use a VPN to access your company system and do not use proxies at all for accessing the internet. Some of your company servers are not accessible directly from the VPN and you should use the hidden proxy to reach them.

So lets put it tougher and write a Sweetcher config:

```yaml
# First lets define our proxies
proxies:
  main: "masterproxy.yourcompany.it:8080"
  hidden: "hiddenproxy.yourcompany.it"

# Then lets define some profiles
profiles:
  atCompany:
    # A profile should have a default proxy if none of its rules match
    default: main
    # Rules are ordered 
    rules:
      - host_wildcard: "gist.github.com"
        proxy: hidden
      - host_wildcard: "*.yourcompany.it"
        # direct is a reserved word that means: "forward the request directly to the targeted site without using a proxy"
        proxy: direct
      - host_wildcard: "*.google.*"
        proxy: hidden
      - host_wildcard: "*.reedit.*"
        proxy: hidden
  homeworking:
    default: direct
    rules:
      - host_wildcard: "someplace.yourcompany.it"
        proxy: hidden

# Finally lets set the current profile
server:
  profile: atCompany
  # setup the listening address
  address: "127.0.0.1:8080"
```

Then all you need to do is to setup 127.0.0.1:8080 as your default proxy for your whole system (ie for gnome, apt, docker and so on).

When you are at home simply change the current profile to `homeworking` and reload sweetcher. All your apps will use the new set of rules.

## Disclaimer

An important part of the proxy package is copied from the excellent https://github.com/elazarl/goproxy/ project
all the credit goes to @elazrd. I only made some adaptations to dynamically set (or not) an http proxy for CONNECT operations (used for HTTPS connections). I was not able to do it with the
original goproxy library. I also plan to add the ability to match on URL patterns not only on host wildcards which is currently not possible for CONNECT operations on goproxy.

## Roadmap

- [ ] Support https proxies in case of HTTPS CONNECT connections
- [x] Dynamic configuration reload
- [ ] URL patterns
- [ ] metering (errors, rate, ...)
- [ ] proxies load balancing
- [ ] Management API (?)


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Floicalbertin%2Fsweetcher.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Floicalbertin%2Fsweetcher?ref=badge_large)