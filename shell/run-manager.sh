COUNT=0
MAX_RETRIES=10
CURRENT_DATE=$(date +'%Y_%m_%d')

PATH_TO_CONFIG="./../cmd/manager/config/config.toml"
PATH_TO_MANAGER="./../cmd/manager/manager"

while [ ${COUNT} -lt ${MAX_RETRIES} ]
do
  CURRENT_LOGS_FILE="logs_${CURRENT_DATE}_$(( COUNT+1 )).txt"

  ${PATH_TO_MANAGER} -config ${PATH_TO_CONFIG} -type "reindex" > "${CURRENT_LOGS_FILE}"

  ERROR_OUTPUT=$(grep ERROR "${CURRENT_LOGS_FILE}" )

  if [ -z "${ERROR_OUTPUT}" ]
  then
    break
  else
    echo "Something went wrong error: ${ERROR_OUTPUT}"
    echo "Will retry "$(( COUNT+1 ))""
  fi

  COUNT=$(( COUNT+1 ))
done

if [ ${COUNT} -eq $(( MAX_RETRIES )) ]
then
  echo "Reindex process failed, check logs file ${LOGS_FILE}"
else
  echo "Reindex accounts with stake success"
fi


