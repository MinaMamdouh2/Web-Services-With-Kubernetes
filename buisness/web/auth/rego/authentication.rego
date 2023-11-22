package rego

# Define a variable auth and set it to false
default auth = false

# Here we are saying we can change auth based on this variable 'jwt_valid'
auth {
	jwt_valid
}

# jwt_valid is assigned based on calling verify_jwt function
jwt_valid := valid {
	[valid, header, payload] := verify_jwt
}

# This has a fuinction that is part of rego 'decode_verify'
# and you can pass inputs as token, public key & iss information
# then it will return verify_jwt and we gonna divide it into 3 parts
verify_jwt := io.jwt.decode_verify(input.Token, {
	"cert": input.Key,
	"iss": input.ISS,
})
