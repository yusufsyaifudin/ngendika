# NGENDIKA [Work In Progress]

> In Javanese, ["ngendika"](https://id.wikipedia.org/wiki/Kata_krama_inggil) means "said".
> i.e: "Ibu ngendika kowé kudu sekolah." in English: "Mother said you must go to school."

Ngendika is a multi-tenant, scalable, and high-performant Push Notification Server.

## Background

During my professional career as Software Engineer, I always found that company tend to write their own Push Notification server.
Some companies may write only single use case Push Notification server, so when another application wants to send PN, it must spin up another server to handle PN.
This because different application/service has different FCM Server Key, APNS key, e.t.c.
Some companies don't use separate PN server, so every PN will be handled by the application itself.

Then, why don't we create one PN service that can be used by any service?
In this "Ngendika" we don't need to separate PN Server to handle development and production, 
because we will see this Ngendika to be used by SaaS company where their client can test their application using Dev or Prod APNS certificate (FCM don't distinguish this).
Or in simply, we can assume that this "Ngendika" is the 3rd party SaaS service that deployed on-premises in your service architecture.

## Installation

* Download pre-built binary from Release Page. 
* Create PostgreSQL version 12+ database.
* Prepare Redis instance.
* Copy config on `config.sample.yaml` to `config.yaml` and modify the value.
* Run migration by running `ngendika -c config.yaml migrate appRepo up` in terminal.
* Hit the API using Postman.

## Features

* [x] FCM Multicast using new API
* [x] FCM using Legacy Message Payload
* [ ] APNS both dev and production
* [x] Webhook HTTP

