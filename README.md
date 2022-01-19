# GoSpam

Simple selfhosted throw-away email

## Screenshots

### Welcome / Mailbox search
![Welcome View](/screenshots/gospam_welcome.png)

### E-Mail Details
![E-Mail Details View](/screenshots/gospam_email.png)


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
AcceptSubdomains: false
MaxStoredMessages: 100000
RetentionHours: 4
CleanupPeriod: 5
MaximumMessageSize: 5242880
SMTPTimeout: 60
MaxRecipients: 10
RandomAliasPlaceholder: false
```

`AcceptedDomains` can be omitted to accept any recipient domain, this will make the server look like an open relay.

To add persistence to the service you can optionally supply connection details for a Redis server:

```
RedisBackend:
  Address: "127.0.0.1:6379"
  Password: ""
  DB: 0
```

The `MaxStoredMessages` does not apply to the redis backend. Additionally searching for aliases using `*` and `?` placeholders 
is available only when using the redis backend.

## Setup

Just run the `gospam` binary.

You have to create DNS records to configure the machine running GoSpam as the mail server for any
domain or subdomain.

The web interface can be configured to listen on a specific interface only (e. g. `127.0.0.1:80` to listen on localhost) 
this makes it possible to make the SMTP listener publicly accessible and keep the web interface private.
