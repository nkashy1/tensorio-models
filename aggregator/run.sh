python -u -m aggregator_run \
  --aggregation-type ${AGGREGATION_TYPE} \
  --ckpt-paths-file ${AGGREGATION_CKPTS_FILELIST} \
  --resource-path ${AGGREGATION_RESOURCE_PATH} \
  --output-path ${AGGREGATION_OUTPUT_PATH} \
  --output-resource-path ${AGGREGATION_OUTPUT_RESOURCE_PATH} \
  --repository ${REPOSITORY} \
  --token ${TENSORIO_MODELS_TOKEN} \
  --export-type ${AGGREGATION_EXPORT_TYPE}