# Data Persistence Challenge

A fun and simple coding challenge to fetch student data partitioned across multiple databases. 

The app uses sqlite databases, which are included in the repo (enrollment1.db & enrollment2.db). The databases have sample data in them. 

Review the QnA.txt file for insight into some decisions made in the app. 

## To Build App

The app is built using Make. Simply run the following command in the root of the project...

```sh
make
```

## To Run App

```sh
./enrollment
```

### To rebuild databases & populate with sample data

```
./enrollment -build_db=true
```