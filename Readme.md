# Svalin

An Open-Source RMM-Software focused on ease of use and security.

## :warning: **WARNING** :warning:

- This code is not yet fit for production!!!
- The cryptographic implementation did not get any proper review.
- The system might still be quite unstable
- Besides the cryptography there is no proper logging or defense mechanism yet.

## Goals

The system should be usable by IT-Admins which arent all that familiar with linux or things like SSH.

Hardening is a bit of an antipattern - everything should be as secure as possible by default

Working with Svalin should be painless, fast and easy.

### Technical goals

These goals are not reached yet, but are intended to guide the development to eventually reach them.

- Setup and integration of new agents should be dead simple
- The system should be secure, even if the server is taken over.

## Project Name

The name "Svalin" comes from nordic mythology. It's the name of a powerfull Shield which holds back the heat of the sun.
It is mounted on the charriot of the goddess Sol and keeps the Head at bay.

Just like it's powerfull name-giver, this software should shield you from the everyday heat you might experience.
It should protect you from cyber-attacks and aid you in supporting your colleagues and customers. 

## Road to 0.1

These systems should be in place before publishing the very first version.

- [ ] Multiuser
- [ ] TCP-Passthrough
- [ ] Revocations
- [ ] Shell
- [X] Signature Chain verification
- [ ] Basic Permission Check (agent shouldnt be able to upload users)
- [X] End to End encryption
- [ ] Nice Logo

## internals

### config

This package provides a basic way to open application data using profiles and reading config keys.

### db

This is a small wrapper around boltDB.
It exposes Scopes, which automatically navigate to their corresponding bucket when viewing or updating data.

### pki

Here you can find the cryptographic code for handling and verifiying certificates as well as signing and verifying data.

### rmm

as the name suggests - you'll find the code for monitoring and interacting with an agent here

### rpc

implements the badly made rpc protocol over quic.
has ready made commands for forwarding and encrypting sessions.

### system
This package houses the more general code which coordinates all functional packages
The sub-packages implement the server, client and agent respectively

### ui
a badly made ui based on the quite nice fyne ui framework.

### util
houses everything that didn't get it's own package and might be useful elsewhere too.

