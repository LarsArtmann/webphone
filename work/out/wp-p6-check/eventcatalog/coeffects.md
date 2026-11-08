# Coeffect Validation Summary

Producer/consumer graph over declared events (derived from service sends/receives; explicit declarations win).

| Event | Producers | Consumers | Status |
| ----- | --------- | --------- | ------ |
| message.received | webphone |  | unconsumed (advisory) |
| message.delivery.advanced | webphone |  | unconsumed (advisory) |
| thread.flag.changed | webphone |  | unconsumed (advisory) |
| fax.advanced | webphone |  | unconsumed (advisory) |

Validation: no dangling subscriptions
