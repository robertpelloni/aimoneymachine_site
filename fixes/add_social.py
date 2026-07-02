import sys

with open(
    "/opt/aimoneymachine/src-tmp/orchestrator/cmd/orchestrator/main.go", "r"
) as f:
    lines = f.readlines()

new_tasks = [
    '\t\tscheduler.Register("SocialTwitter", 6*time.Hour, func(o *orchestrator.Orchestrator) error {\n',
    '\t\t\treturn protocol.HandleURI("hustle://social?platform=Twitter&topic=AI+automation")\n',
    "\t\t})\n",
    '\t\tscheduler.Register("SocialLinkedIn", 12*time.Hour, func(o *orchestrator.Orchestrator) error {\n',
    '\t\t\treturn protocol.HandleURI("hustle://social?platform=LinkedIn&topic=AI+automation")\n',
    "\t\t})\n",
]

insert_line = None
for i, line in enumerate(lines):
    if 'scheduler.Register("SwarmSync"' in line:
        insert_line = i
        break

if insert_line is not None:
    result = lines[:insert_line] + new_tasks + lines[insert_line:]
    with open(
        "/opt/aimoneymachine/src-tmp/orchestrator/cmd/orchestrator/main.go", "w"
    ) as f:
        f.writelines(result)
    print(f"Inserted social tasks before line {insert_line + 1}")
else:
    print("ERROR: Could not find insertion point")
    sys.exit(1)
