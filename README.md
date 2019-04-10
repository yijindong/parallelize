# parallelize
PARALLELIZE allows you to run BASH SHELL scripts in parallel.You have the flexibility to specify the number of parallelism, and the parameters.And supports parameter files in CSV format.

We often use this script to complete batch parallel backup and import of ORACLE and MYSQL database tables, and the performance is stable and good.

Usage:
  example, you need run:
-    a.sh 1 1
-    a.sh 2 2
-    a.sh 3 3
-    a.sh 4 4
-    a.sh 5 5
-  to use parallize
-    ./parallize a.sh 1,1 2,2 3,3 4,4 5,5
-    ./parallize -n 2 a.sh 1,1 2,2 3,3 4,4 5,5
-    ./parallize -csv a.csv -n 2 a.sh
-    a.csf:
-    1,1
-    2,2
-    3,3
-    4,4
-    5,5

install:
   go build main.go
  

[mysql@collie1 ~]$ ./parallelize_linux  -h
Usage of ./parallelize_linux:
  -csv string
        CSV file containing arguments for each job
  -n workers
        Number of workers (default 8)
  -q    Quiet

