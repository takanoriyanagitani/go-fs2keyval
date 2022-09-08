## Benchmark Example

#### Write amplification check

- data size: 8 KiB
- SSD 1: (NVMe) Seagate FireCuda 530
- SSD 2: (SATA) Intel S4510/S4610/S4500/S4600 Series

| SSD # | iterations | kv pairs / iteration | data units written | KiBW / iteration | WA           |
|:-----:|:----------:|:--------------------:|:------------------:|:----------------:|:------------:|
| 1     | 1024       | 1                    | 100 ~ 120          | 53 ~ 55          | 660% ~  690% |
| 1     | 1024       | 10                   | 350 ~ 400          | 17 ~ 20          | 210% ~  250% |
| 2     | 1024       | 1                    | 130 ~ 200          | 60 ~ 100         | 800% ~ 1200% |
| 2     | 1024       | 10                   | 390 ~ 400          | 19 ~ 20          | 230% ~  240% |
