jq  $' 
  .[] | select(.id=="transformation") |
    .configurations[] |
    . as $config |
    .rows[] |
    {  name, queries: .configuration.queries }' "./data/10_Datamart/03_Components.json" -r
