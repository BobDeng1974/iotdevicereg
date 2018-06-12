octet = require 'octet'
ecdh = require 'ecdh'
json = require 'json'

keyring = ecdh.new('ec25519')
keyring:keygen()

output = json.encode({
	public = keyring:public():base64(),
	secret = keyring:private():base64()
})

print(output)