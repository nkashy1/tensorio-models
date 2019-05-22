import aggregator




MODEL_DIRECTORY_NAMES_LOCAL = [
    "/tmp/federated_checkpoints/local_update_1",
    "/tmp/federated_checkpoints/local_update_2"
]
MODEL_DIRECTORY_NAMES_GCS = [
    "gs://path/to/gcs/local_update_1",
    "gs://path/to/gcs/local_update_2"
]
OUTPUT_PATH_CMA_LOCAL = "/tmp/federated_checkpoints/aggregated-cma"
OUTPUT_PATH_WCMA_LOCAL = "/tmp/federated_checkpoints/aggregated-wcma"
OUTPUT_PATH_CMA_GCS = "gs://path/to/gcs/aggregated-cma"
OUTPUT_PATH_WCMA_GCS = "gs://path/to/gcs/aggregated-wcma"


def local_test():
    print ("Testing CMA Aggregator: ")
    agrgtr_cma = aggregator.Aggregator(aggregator_type="cma", debug=True)
    agrgtr_cma.aggregate(MODEL_DIRECTORY_NAMES_LOCAL, OUTPUT_PATH_CMA_LOCAL)

    print ("Testing WCMA Aggregator: ")
    agrgtr_wcma = aggregator.Aggregator(aggregator_type="wcma", debug=True)
    agrgtr_wcma.aggregate(MODEL_DIRECTORY_NAMES_LOCAL, OUTPUT_PATH_WCMA_LOCAL)
    for i in range(4):
        # Tests that sessions and default graphs are reset
        agrgtr_wcma.aggregate(MODEL_DIRECTORY_NAMES_LOCAL, "{0}-{1}".format(OUTPUT_PATH_WCMA_LOCAL, i))


def gcs_test():
    print ("Testing CMA Aggregator: ")
    agrgtr_cma = aggregator.Aggregator(aggregator_type="cma", debug=True)
    agrgtr_cma.aggregate(MODEL_DIRECTORY_NAMES_GCS, OUTPUT_PATH_CMA_GCS)

    print ("Testing WCMA Aggregator: ")
    agrgtr_wcma = aggregator.Aggregator(aggregator_type="wcma", debug=True)
    agrgtr_wcma.aggregate(MODEL_DIRECTORY_NAMES_GCS, OUTPUT_PATH_WCMA_GCS)
    for i in range(4):
        # Tests that sessions and default graphs are reset
        agrgtr_wcma.aggregate(MODEL_DIRECTORY_NAMES_GCS, "{0}-{1}".format(OUTPUT_PATH_WCMA_GCS, i))


if __name__ == '__main__':
    local_test()
    gcs_test()

