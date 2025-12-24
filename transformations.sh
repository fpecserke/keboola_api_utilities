SEARCH="${1:-CUSTOMER_ATTR}"

echo "Searching for '$SEARCH' ..."
for dir in ./data/*/; do
  echo "Processing: $dir"
  PROJECT="$dir"
  PROJECT_ID=$(jq $' .owner.id' "${PROJECT}00_Verify_Token.json")

echo "project ID: $PROJECT_ID"

# OLD LEGACY TRANSFORMATIONS
jq --arg pid "$PROJECT_ID" --arg search "$SEARCH" $' 
  .[] | select(.id=="transformation") |
    .configurations[] |
    . as $config |
    .rows[] |
    . as $row |
    .configuration | {
      name,
      input: .input // [] | map({source, destination} | select(.source | contains($search))),
      link: "https://connection.eu-central-1.keboola.com/admin/projects/"+$pid+"/legacy-transformation/bucket/"+$config.id+"/transformation/"+$row.id
    } | select(.input!=[]) |

       "**OLD LEGACY** ["+$config.name + " | " + $row.name + " | " + .name +"]\n("+ .link + ")\n\n"
     
    ' "${PROJECT}03_Components.json" -r


# NEW LEGACY TRANSFORMATIONS
jq --arg pid "$PROJECT_ID" --arg search "$SEARCH" $' 
  .[] | select(.id=="keboola.legacy-transformation") | 
  .configurations // [] | .[] | 
  . as $config |
  .configuration.parameters.transformations // [] | .[] | 
  .parameters | { 
    name,
    input: .input // [] | map({source} | select(.source | contains($search))),
    link: "https://connection.eu-central-1.keboola.com/admin/projects/"+$pid+"/legacy-transformation/bucket/"+$config.id+"/transformation/" + .id
  } | 
  select(.input != []) | 

  "**NEW LEGACY** ["+$config.name + " | " + .name +"]\n("+ .link + ")\n\n"
  
  ' "${PROJECT}03_Components.json" -r

  # TRANSFORMATIONS VERSION 2 - latest
  jq --arg pid "$PROJECT_ID" --arg search "$SEARCH" $' 
  .[] | select(.id=="keboola.snowflake-transformation") |
  .configurations[] |
  . as $conf |
  .configuration |
  {
    name,
    input: .storage.input.tables // [] | map({source, destination} | select(.source | contains($search))),
    url: "https://connection.eu-central-1.keboola.com/admin/projects/"+$pid+"/transformations-v2/keboola.snowflake-transformation/"+$conf.id,
  } |
  select(.input!=[]) |

  "**TRANSFORMATION V2** [" + $conf.name + "]\n(" + .url + ")\n\n"
  
  ' "${PROJECT}03_Components.json" -r

done
