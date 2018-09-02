# UnFail2Ban
Web front-end and back-end to unban those hit by Fail2Ban. Written for UCC Netsoc. Coded by Noah Santschi-Cooney (Strum355)

## Dependencies

- Golang
- TypeScript

## Requirements

Needs to be able to connect to ip-api.com.
Uses LDAP for auth.
>More details to be added later

## Features

- View and unban IPs
- Login via LDAP
- ~~Fail2Ban log viewing~~ returning soon

## Install

- Config is expected to be at `/etc/unfail2ban/settings.conf`. TODO env var for development
- `static` folder location is defined in the config
- To run, run `run.sh`. To build, run `build.sh`

Sample:

```toml
jail="sshd"
port="8080"
cookie_host="uf2b.domain.com"
listen_host="uf2b.domain.com"
file_dir="/etc/unfail2ban/"
LDAP_Key="banana"
LDAP_Host="ldap.domain.com:389"
LDAP_User="cn=admin,dc=domain,dc=com"
LDAP_BaseDN="dc=domain,dc=com"
```

## TODO

- Issues
- CI setup