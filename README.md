# chorus

Step 1 - Create the MVP which is a single server, no other tools that allows users to connect and get put in a lobby that is managed by a  javascript.  That script can then do matchmaking to put you in a tic-tac-toe game, a room managed by another javascript.

This works well enough.

Step 2 - Route everything through RedPanda/Kafka.  How does this change things

- Create an EndUser server that let's users connect and provides access to the system.
    - Whe a user connects they join the GlobalLobby 
        - To Join a Room
            - A record keyed by the roomid is updated to have the connection id
            - We start consuming the rooomid topic 

            
I've played around with this a bit and it changes the design significantly, as expected.

Step 3 - Switched to using Redis and this had other issues.  I did have fun making distributed data structures.


Step 4 - Now another set of changes
    - Postgres for persistance
    - RedPanda For sending messages 

What are our major structures:
    - Machines 
        - End User Server 
            - Accepts connections from users
            - Default joins Global Lobby
        - Room Server
            - runs room logic
    - User Connections
        - This object connects to a user talking into the system and handles all communication 
    - Room
        - Attached to a script context for room logic
        - Has users joined to it
    - Monitor - single machine in the cluster that orchestrates things

We use Postgres to store state, this includes
    - Current connections
    - Room Memberships
    - Machine status

We use RedPanda to send messages between machines
    - Room Channels - each room has a channel
	- all room events go through this channel
    - Machine Channels - each machine has a channel
	- join - tells a user/connection to join a room
	- leave - tells a suer/connection to leave a room

What do we need to do?
    done - Postgres docker instance
    done - scripts for creating DB
    - Postgres library
    - Kafka library

