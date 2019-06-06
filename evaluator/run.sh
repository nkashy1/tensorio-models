python -m evaluator_run \
       --ckpt-path ${EVALUATION_CKPT_PATH} \
       --npy-file-0class ${EVALUATION_0CLASS_NPY} \
       --npy-file-1class ${EVALUATION_1CLASS_NPY} \
       --output-file ${EVALUATION_OUTPUT_JSON}
