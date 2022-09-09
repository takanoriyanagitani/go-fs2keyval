# go-fs2keyval
filesystem as key value store

[![Go Reference](https://pkg.go.dev/badge/github.com/takanoriyanagitani/go-fs2keyval#FileLike.svg)](https://pkg.go.dev/github.com/takanoriyanagitani/go-fs2keyval#FileLike)
[![Go Report Card](https://goreportcard.com/badge/github.com/takanoriyanagitani/go-fs2keyval)](https://goreportcard.com/report/github.com/takanoriyanagitani/go-fs2keyval)
[![codecov](https://codecov.io/gh/takanoriyanagitani/go-fs2keyval/branch/main/graph/badge.svg?token=OGAA9OIPWV)](https://codecov.io/gh/takanoriyanagitani/go-fs2keyval)

```

DB Instances:

  - instance 1:
  - instance 2:
  - instance 3:
  - ...
  - instance x:
    options:
      fs:
        instance root dir: /path/to/instance/x/databases
    databases:
      - database 1:
      - database 2:
      - database 3:
      - ...
      - database y:
        options:
          fs:
            database root dir: ${instance root dir}/y/buckets
        buckets:
          options:
            fs:
              type: tar
              filename: ${database root dir}/cafef00ddeadbeafface864299792458.tar
              contents:
                - data_2022_09_09_f00ddeadbeaffacecafe864299792458/bucket.txt
                - data_2022_09_09_f00ddeadbeaffacecafe864299792458/key.bin
                - data_2022_09_09_f00ddeadbeaffacecafe864299792458/val.bin
                - data_2022_09_09_deadbeaffacecafef00d864299792458/bucket.txt
                - data_2022_09_09_deadbeaffacecafef00d864299792458/key.bin
                - data_2022_09_09_deadbeaffacecafef00d864299792458/val.bin
                - ...

Directory tree

/path/to/instance/0
/path/to/instance/1
/path/to/instance/2
/path/to/instance/...
/path/to/instance/x
  |
  +-- databases/0
      databases/1
      databases/2
      databases/...
      databases/y
        |
        +-- cafef00ddeadbeafface864299792458.tar
              |
              +-- data_2022_09_09_f00ddeadbeaffacecafe864299792458/bucket.txt
                  data_2022_09_09_f00ddeadbeaffacecafe864299792458/key.bin
                  data_2022_09_09_f00ddeadbeaffacecafe864299792458/val.bin
                  data_2022_09_09_deadbeaffacecafef00d864299792458/bucket.txt
                  data_2022_09_09_deadbeaffacecafef00d864299792458/key.bin
                  data_2022_09_09_deadbeaffacecafef00d864299792458/val.bin

```
