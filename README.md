# driusan's DKIM tools

This is a collection of pure Go tools I've written to DKIM sign
messages and verify them on my 9front mail server.  They should be
fairly easy to incorporate into any pipeline that can pass messages
along stdin and read them from stdout.

## Verifying DKIM Signatures

The tool `dkimverify` will verify that a message has a valid
signature.

It can either read a message from stdin, or have (optionally many)
filenames passed as arguments.  If all messages have valid signatures,
it will exit with a success status, otherwise it will exit with an
exit code of the number of messages that failed validation.  For each
one, it will print the reason for the failure to stderr.

The `-hd` parameter takes a string argument and instead of printing to
stderr, will print an SMTP header of that name with a value of "Pass"
or "Fail" to stdout.  Temporary failures or no DKIM signature present
in a message will print nothing.  `-hdprefix` or `hdsuffix` can be
used to add a prefix or suffix to the header value.

The dkimverify tool can be used without any special configuration.

## Signing DKIM Signatures

Signing is slightly more complicated by necessity.  The tool
`dkimkeygen` will create 2 files in the directory it's run in:
`dns.txt` and `private.pem`.  The contents of the dns.txt need to be
added to your domain's DNS as a TXT record at
`selector._domainkey.example.com` so that the DKIM signatures added by
`dkimsign` can be validated.  (the "selector" part can be anything you
want, but needs to match what's passed to `dkimsign`) `private.pem` is
the corresponding private key.

`dkimsign` reads a message from stdin and writes a signed version of
that message to stdout according to the parameters passed.  The
incoming message can have any line ending, but they'll be converted to
"\r\n" line endings on output (unless `-n` is passed, in which case
they'll be printed as "\n").  If "-hd" is passed to `dkimsign`, the
header will be printed to stdout but not the message body.  The `-key`
parameter is the private key and should be the path to the
`private.pem` generated by dkimkeygen.  `-s` is the selector and
should match the selector part of the domain name.  `-d` is the domain
name.

### Example (Plan 9)

The following is an example `/mail/lib/remotemail` that should add
valid DKIM Signatures on a Plan 9 server.  Note that since it requires
the sender to have access to private.pem it should probably only be
used in a single-user environment.

```
#!/bin/rc
shift
sender=$1
shift
addr=$1
shift

# The first smtp with -f causes SMTP to add any headers it wants and
# fully qualify the addresses, printing to stdout instead of sending.
# dkimsign signs From, Date, Subject, and To headers (and the body)
# with relaxed encoding (the default).  It prints the signed message
# to stdout with \n line endings ("-n") and undoes the dot-stuffing
# added by smtp -f ("-u") both for generating the signature and for
# printing it.  It uses the selector 19700101._domainkey.example.com and
# signs using the private key at /sys/lib/dkim/private.pem (which
# should correspond to the selector)
# The second smtp (without -f) then sends the mail for real.
exec /bin/upas/smtp -f -h example.com .example.com $addr $sender $* | /bin/dkimsign -h From:Date:Subject:To -u -s 19700101 -n -d example.com -key /sys/lib/dkim/private.pem | /bin/upas/smtp -s -h example.com .example.com $addr $sender $*
```
