# chorus

Step 1 - Create the MVP which is a single server, no other tools that allows users to connect and get put in a lobby that is managed by a  javascript.  That script can then do matchmaking to put you in a tic-tac-toe game, a room managed by another javascript.

This works well enough.

Step 2 - Route everything through RedPanda/Kafka.  How does this change things

- Create an EndUser server that let's users connect and provides access to the system.
    - Whe a user connects they join the GlobalLobby 
        - To Join a Room
            - A record keyed by the roomid is updated to have the connection id
            - We start consuming the rooomid topic 

            