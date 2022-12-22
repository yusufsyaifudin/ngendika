---
title: Ngendika
---


Ngendika is a self-hosted notification server that easy to use.

{{< mermaid >}}
%%{init:{"theme":"forest"}}%%
graph LR;
I[Install] --> App[Create App]


App -->|Email| EmailCfg[Save Email Config]
EmailCfg --> SendEmail[Send Email]

App -->|FCM| FCMKey[Save FCM Service Account Key]
FCMKey --> SendFCM[Send Push Notification through FCM]

App -->|Webhook| Web[Webhook URL]

{{< /mermaid >}}


Using Ngendika, you doesn't need to write or deploy the Push Notification server each time you need it.

Think of it as a Push Notification SaaS but you own your infrastructure!

Ngendika is:

* Scalable - Each configuration of an Application (or tenant) is saved in the same table, but we use partitioning to ensure the data is scoped.


{{% button href="/getting-started" %}}Getting Started{{% /button %}}