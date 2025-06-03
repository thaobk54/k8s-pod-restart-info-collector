# k8s-pod-restart-info-collector

k8s-pod-restart-info-collector is a simple K8s custom controller that watches for Pods changes and collects K8s Pod restart reasons, logs, and events to **Slack or Discord** when a Pod restarts.

---

## 🚨 Now Supports Discord Notifications!

You can now send pod restart alerts to either Slack **or** Discord. The notification platform is selected via the `NOTIFIER_TYPE` environment variable.

### How to Use Discord Notifications

1. **Set Environment Variables:**
   ```sh
   export NOTIFIER_TYPE=discord
   export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your_webhook_id/your_webhook_token"
   export CLUSTER_NAME="your-cluster"
   # Optional:
   export DISCORD_USERNAME="k8s-pod-restart-info-collector"
   export MUTE_SECONDS=600
   ```
2. **Run the Controller:**
   ```sh
   go run .
   ```
3. **Result:**
   - Pod restart alerts will be sent to your Discord channel via the webhook.

> If `NOTIFIER_TYPE` is not set or is set to `slack`, the controller will use Slack notifications (see below for legacy instructions).

---

## Overview of the Data Collected

Here are two Slack screenshots of the example messages (Discord messages are formatted similarly):

### Brief Alert Message
![image](https://miro.medium.com/max/1200/1*iFQeWKHZv3zzJC8lgiZtjA.png)

### Detailed Alert Message

As shown below, by clicking "Show more", we can see the Reason, "Pod Status", "Pod Events", "Node Status and Events", and "Pod Logs Before Restart".

![image](https://miro.medium.com/max/1200/1*mvzXhbNeQCJ9Blh1oDH4uw.png)

---

## How to test and develop locally (Slack Example)

```bash
export SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxxxx/xxxxx
export NOTIFIER_TYPE=slack
# ... other env vars ...
go run .
```

## Install using Helm

**Replace the `slackWebhookUrl`, `clusterName` and  `slackChannel`.**

```bash
helm diff upgrade --install k8s-pod-restart-info-collector ./helm \
   --set slackWebhookUrl="https://hooks.slack.com/services/Change-Me" \
   --set clusterName="Change-Me" \
   --set slackChannel="Change-Me"
```

Check Commands:

```bash
# check commands
kubectl get pod,deploy,sa,secret -l app.kubernetes.io/instance=k8s-pod-restart-info-collector
helm status k8s-pod-restart-info-collector
helm get values k8s-pod-restart-info-collector
helm get manifest k8s-pod-restart-info-collector
helm get all k8s-pod-restart-info-collector
# see logs
kubectl logs deployment/k8s-pod-restart-info-collector -f
```

Run a `debug-pod` to verify the collector:

```bash
kubectl run debug-pod --image=alpine -- date;sleep 30
kubectl get pod debug-pod -w
```

## Uninstall

To uninstall/delete the `k8s-pod-restart-info-collector` helm release:

```bash
helm uninstall k8s-pod-restart-info-collector
```

> The command removes all the Kubernetes components associated with the chart and deletes the release.

## Helm Parameters

| Name                                | Description                                        | Value         |
| ------------------------------------| -------------------------------------------------- | ------------- |
| `clusterName`                       | K8s cluster name (Display on notification)         | required      |
| `slackUsername`                     | Slack username (Display on slack message)          | default: `"k8s-pod-restart-info-collector"` |
| `slackChannel`                      | Slack channel name                                 | default: `"restart-info-nonprod"` |
| `muteSeconds`                       | The time to mute duplicate pod alerts              | default: `"600"` |
| `ignoreRestartCount`                | The number of pod restart count to ignore          | default: `"30"` |
| `ignoredNamespaces`                 | Namespaces to be ignored (comma-separated/regex)   | default: `""` |
| `ignoredPodNamePrefixes`            | Pod name prefixes to be ignored (comma-separated/regex) | default: `""` |
| `watchedNamespaces`                 | Namespaces to be watched (comma-separated/regex)   | default: `""` |
| `watchedPodNamePrefixes`            | Pod name prefixes to be watched (comma-separated/regex) | default: `""` |
| `ignoreRestartsWithExitCodeZero`    | Ignore restart events with exit code 0             | default: `false` |
| `slackWebhookUrl`                   | Slack webhook URL                                  | required if slackWebhooUrlSecretKeyRef is not present |
| `slackWebhookurlSecretKeyRef.key`   | Slack webhook URL SecretKeyRef.key                 | |
| `slackWebhookurlSecretKeyRef.name`  | Slack webhook URL SecretKeyRef.name                | |

### Discord Environment Variables
| Name                  | Description                                 | Example/Default |
|-----------------------|---------------------------------------------|-----------------|
| `NOTIFIER_TYPE`       | Set to `discord` to enable Discord alerts   | `discord`       |
| `DISCORD_WEBHOOK_URL` | Discord webhook URL (required)              |                 |
| `DISCORD_USERNAME`    | Bot username (optional)                     | `k8s-pod-restart-info-collector` |
| `CLUSTER_NAME`        | Cluster name (for message context)          |                 |
| `MUTE_SECONDS`        | Mute duplicate alerts (optional)            | `600`           |

## FAQ

1. When will the collector send Pod restart messages to Slack or Discord?

   When a Pod restarts. However, if one of the following conditions is met, the messages are not sent.
   1. Pod restartCount > 30
   2. In the previous 10 minutes, the same Pod restart message was sent

2. How to customize notification channel for each pod (Slack only)

   Adding `alert-slack-channel: "your-slack-channel-name"` to Pod annotations or labels.
   For example, a label: `alert-slack-channel: "restart-info-nonprod"`

---

## How to write a K8s controller
Please refer to:
- https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md
- https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
- https://github.com/kubernetes/client-go/tree/master/examples/workqueue

## Copyright and license

Copyright [2022] [Airwallex (Hong Kong) Limited]

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Contribution

If you are interested in contributing, see [CONTRIBUTION.md](./CONTRIBUTION.md).
