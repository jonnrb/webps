/*
Package psk allows using TLS using certificates generated on the fly from
pre-shared keys. This is great for situations where a full-blown PKI is
overkill, but you need the stability of TLS.

NOTE: I'm overloading PSK here. These are asymmetric keys, but the principle is
the same.
*/
package psk // import "go.jonnrb.io/webps/psk"
