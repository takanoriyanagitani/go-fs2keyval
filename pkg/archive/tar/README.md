## Benchmark Example

#### Write amplification check

- data size: 8 KiB
- SSD 1: (NVMe) Seagate FireCuda 530
- SSD 2: (SATA) Intel S4510/S4610/S4500/S4600 Series
- SSD 3: (NVMe) Intel SSDPEKKW010T7
- SSD 4: (SATA) SanDisk SDSSDH31000G
- SSD 5: (NVMe) ADATA APSFG-2TCS

| SSD # | iterations | kv pairs / iteration | data units written\* | KiBW / iteration | WA            |
|:-----:|:----------:|:--------------------:|:--------------------:|:----------------:|:-------------:|
| 1     | 1024       | 1                    |  100 ~   120         |  53 ~  55        |  660% ~  690% |
| 1     | 1024       | 10                   |  350 ~   400         |  17 ~  20        |  210% ~  250% |
| 2     | 1024       | 1                    |  130 ~   200         |  60 ~ 100        |  800% ~ 1200% |
| 2     | 1024       | 10                   |  360 ~   420         |  17 ~  21        |  220% ~  260% |
| 3     | 1024       | 1                    |   90 ~   100         |  47 ~  48        |  590% ~  600% |
| 3     | 1024       | 10                   |  350 ~   360         |  17 ~  18        |  210% ~  220% |
| 4     | 32768      | 1                    | 4100 ~  6300         |  60 ~ 100        |  800% ~ 1200% |
| 4     | 32768      | 10                   | 9400 ~ 11400         |  14 ~  18        |  180% ~  220% |
| 5     | 1024       | 1                    |  200 ~   230         | 100 ~ 120        | 1200% ~ 1500% |
| 5     | 1024       | 10                   |  470 ~   490         |  24 ~  25        |  290% ~  300% |

###### \*data units written

- SSD 1,3,5
    - using: raw value(1 unit = 512\*1000 bytes)
    - raw value = Data Units Written
- SSD 2
    - using: (raw value) \* (32\*1,048,576)/512,000
    - raw value = HostWrites\_32MiB(ID 241)
- SSD 4
    - using: (raw value) \* (1048576 \* 1024)/512,000
    - raw value = Total\_Writes\_GiB(ID 241)
