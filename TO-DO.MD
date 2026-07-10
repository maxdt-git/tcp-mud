# To-Do
- [OK] command cant be other than connect if not connected
We can add player to other command for easier access.
## GROUP INVITE
- [OK]cant invite itself
- [OK]cant invite player in a group
## GROUP JOIN
- [OK]send event to all group members when player joins
## GROUP LEAVE
- [OK]remove group member when leaving game
## CHAT ALL
- [OK]implement global chat
- [OK]implement room chat
- [OK]implement group chat
## STATUS
- [OK]implement status command and modification to player struct
## LOOK
- []look -> update players when delete
## COMBAT SYSTEM
- []create attackable npcs in json
- []can only attack monsters
- []design combat system
## QUEST SYSTEM
- []QUEST command to accept a quest from npc
- []QUESTS command to list all accepted quests -> modify player struct to accept Quests map[string]string
- []send EVT when quest prerequisites are filled
- []auto-validate versus manual completion by coming back to the quest-giver
- []quest chains -> quest 2 from npc becomes available only if quest 1 is done
- []yo max c'est bien ton week-end?