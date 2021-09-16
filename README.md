# GoSpam

Simple selfhosted throw-away email

## Configuration

GoSpam supports various configuration options. All options with their respective defaults are listed below.
The configuration file is a yaml file and can be located in the current directory (`gospam.conf`) in `etc` (`/etc/gospam/gospam.conf`) or
in the users home directory (`$HOME/.gospam/gospam.conf`).

```
SMTPListenAddress: ":25"
HTTPListenAddress: ":80"
Domain: "localhost"
AcceptedDomains:
  - "spam4.jonaskoeritz.de"
  - "localhost"
MaxStoredMessages: 100000
RetentionHours: 4
CleanupPeriod: 5
MaximumMessageSize: 5242880
SMTPTimeout: 60
MaxRecipients: 10
RandomAliasPlaceholder: false
```

`AcceptedDomains` can be omitted to accept any recipient domain, this will make the server look like an open relay.

## Setup

Just run the `gospam` binary.

You have to create DNS records to configure the machine running GoSpam as the mail server for any
domain or subdomain.

The web interface can be configured to listen on a specific interface only (e. g. `127.0.0.1:80` to listen on localhost) 
this makes it possible to make the SMTP listener publicly accessible and keep the web interface private.
