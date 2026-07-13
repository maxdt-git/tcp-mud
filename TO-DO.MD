# To-Do
- [OK] command cant be other than connect if not connected
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
- [YEAH_trop_bien_les_vacances]yo max c'est bien ton week-end?
- []*et toi?*
## TALK
- [OK]no hardcode on TALK
- [OK]works for first room but not for others, change code should run with any room
- []fix problem when wrong answer of the riddle, continue to try (or not)
- []create test with other rooms
## JSON
- []think about the next step (monsters, how many quest, purpose of the game etc)

## CHANGES SINCE LAST PUSH
- change in models.go to accept struct npc and dialogue node and in config.json
- its like in pacman with different state in the dialogue, either we stay because of a wrong answer or we continue the dialogue if the answer is true until the end of it
- permits to remove in the main.go the command of the 1st room talk npc
- esaier to find every node of the dialogue
