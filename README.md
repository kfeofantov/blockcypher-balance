## Blockcypher testnet balance

### ENVs

`BC_TOKEN` - Blockcypher api key  
`HTTP_ADDR` - HTTP Server Serve addreess

### Example

#### Request

`curl localhost:9091/bitcoin/addresses/balances --data "addresses=34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo,C9LBdupQfLTtgsKDNRdeo6AroDMAeqoEqD"`

#### Reesponse

`{"data":{"C9LBdupQfLTtgsKDNRdeo6AroDMAeqoEqD":200094},"context":{"code":200}}`