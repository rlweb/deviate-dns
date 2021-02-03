# Deviate DNS

Easily setup redirects via a simple TXT record and use our free service.

Setup a redirect via:
1. point your domain or sub-domain to *IPAddress*
2. Add a TXT record with at least the following:
`"v=deviate-dns1 goto:www.google.com email:rhyslaval@gmail.com` There are also these optional extras: ` statuscode:301 keeppath:false`
   

## Todo:
- upgrade to aes_key_secret_id storage