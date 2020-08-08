# Tsatsubii Auth Service

This is just an experiment.

## Prerequisites

In order to install this you will need to have the following:

1. Go
2. PostgreSQL
3. RabbitMQ
4. Docker (optional)

### Go

How to install Go will depend on your operating system. Most operating systems will have Go packages. I'm running Go 1.14.1 on Arch Linux. I can't guarantee that it will work with older versions.

You may want to set the following in ~/.zshrc (or whatever makes sense in your environment):

    export GO111MODULE=on

### PostgresSQL

I've tested with PostgreSQL 12.2 running on Arch Linux. You will need to perform some tasks as the postgres user. So let's start be switching.

   % sudo su - postgres
   $

Notice how the prompt has changed from the normal % user prompt to a $.

First you should make sure that remote connections are allowed.

    $ $EDITOR data/postgres.conf

Find the listen_addresses directive. This is probably going to be set to localhost by default, which is fine if you are going to run tsatubii-auth and PostgreSQL on the same host. Otherwise you will need to change it. On my box, the interface docker0 is set to 172.17.0.1, so that is the interface I will listen to. Save the file and exit from the editor.

Next you going to set up PostgreSQL client configuration.

    $ $EDITOR data/pg_hba.conf

You will have to specify a connection type, which is going to be 'host', a database, a user and address range and a method. The database and user are going to be the ones you go ahead and create in the next step, so make sure they are the same. I am going to specify 172.17.0.0/16 as the address, to allow all my docker containers to connect. Finally, I will use md5 for the method. This gives me the following line to be added to the end of data/pg_hba.conf:

    host ttb_auth ttb_auth 172.17.0.0/16 md5

Before we create a user and database, let's go ahead and restart PostgreSQL, to make sure we didn't make any errors. I'll do that in a different terminal so that I stay logged in as postgres. We will still need that.

    % sudo systemctl restart postgresql

Good, now we can create a new PostgreSQL user. As user postgres:

    $ createuser --interactive -P ttb_auth

Of course, you don't have to call the user ttb_auth, but it needs to match the user you specified in pg_hba.conf. Allow the user to create databases. Set a password. You are done. Now let's test this user and at the same time create a database.

   % psql --user tt_auth template1

You should get a password prompt and then be put into the psql shell. Create a database, then exit the PostgreSQL shell.

   template1=> create database ttb_auth;
   CREATE DATABASE
   template1=> \q

Again, you might want to call your database something else. That is fine, as long as it matches what you entered in pg_hba.conf.

You will need to create the pgcrypto extension in this databaase. This should be done as the postgres user (unless you made your new user a superuser).

    $ psql ttb_auth
    ttb_auth=# create extension pgcrypto;
    CREATE EXTENSION
    #ttb_auth=# \q

Now you are done with the PostgreSQL configuration.

### RabbitMQ

### Docker

## Building

To build the project

  $ go build

## Using

# Acknowledgements

In the absence of a logo, I found an image of an ant at https://www.pngrepo.com/svg/202145/ant, which will do for now. I don't know who created it, but according to the site, it is LICENSE under Creative Commons 4.0.