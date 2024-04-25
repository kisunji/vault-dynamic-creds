# vault-dynamic-creds

vault-dynamic-creds is a CLI tool to fetch a secret and renew it.

I am currently using it for work with hardcoded strings.

## How to use

```shell
# make sure VAULT_ADDR and VAULT_TOKEN are populated

vault-dynamic-creds --service billing --role ro
```