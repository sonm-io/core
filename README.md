MetricStructures system
=====================
This repository is an aggregate of structures, useful for the hub and the miner.
The structures consist of 3 types of metrics : structural, financial and technical.
These structures are a basis for creating a filtering system, which in itself is the basis for the rating system.
The results of the rating system are used in the dataset for the machine learning algorithm, which in turn can serve as a prediction tool for the hub and the miner, easing the choice and raising the trust between the user and the system.

## Structural Metrics

### Hub Structural metrics
1) Hub Address - records the hub address
2) HubPing - uses pings to diagnose the system and figure out the uplink speed between the hub and the source. Determines whether there are any packets lost between the source and the hub.
3) HubService - the amount of services, available to the hub (not yet implemented).
4) CreationDate - date, on which the hub was activated (registered). Determines for how long the hub was been registered (in order to assess the activity levels for that period).
5) HubLifeTime - determines the lifetime of the hub.
6) PayDay - this attribute sets the amount of money that the hub can pay out
7) FreezeTime - the overall amount of time the hub spent being frozen.
8) AmountFreezeTime - how many times the hub was frozen.
9) TransferLimit - this function sets the transfer limit for the hub.
10) SuspectStatus - this status becomes true if the hub is suspected to be involved with fraud.
11) DayLimit - the limit on the amount of money that the hub can send.
12) AvailabilityPresale - attribute that shows whether or not the hub has presale tokens.
13) SpeedConfirm - this sttribute determines the response time for the hub, which in turn influences the activity probability for the hub.
14) HubStack - the attribute determines how much are the participants holding in their wallets.

### Miner Structural metrics
1) MinAddress - records the miner address.
2) MinPing - uses pings to diagnose the system and figure out the uplink speed between the miner and the source. Determines whether there are any packets lost between the source and the miner.
3) MinStack - the attribute determines how much are the participants holding in their wallets.
4) CreationDate - date, on which the miner was activated (registered). Determines for how long the miner was been registered (in order to assess the activity levels for that period).
5) MinService - the amount of services, available to the hub (not yet implemented).

## Technical Metrics

## Financial Metrics 