import json
import os

from dotenv import load_dotenv

from esclient import ElasticClient


def main():
    load_dotenv()
    aliases = json.loads(os.getenv('ALIASES'))
    print(aliases)

    es_client = ElasticClient(os.getenv('ES_URL'), os.getenv('ES_USERNAME'), os.getenv('ES_PASSWORD'))

    for alias in aliases:
        indices = sorted(es_client.get_indices_behind_alias(alias))
        print(indices)
        if len(indices) == 1:
            compare_indices(es_client, indices[0],"")
            continue
        for idx in range(len(indices)-1):
            compare_indices(es_client, indices[idx], indices[idx+1])


def compare_indices(es_client: ElasticClient, index1: str, index2: str):
    num_ids = int(os.getenv("NUM_ELEMENTS_TO_CHECK_IN_INDEX"))
    ids_idx1 = es_client.get_ids_from_index(index1, num_ids, sort="desc")
    ids_idx2 = es_client.get_ids_from_index(index2, num_ids, sort="asc")

    duplicated_ids = {}
    for doc_id_1 in ids_idx1:
        if doc_id_1 in ids_idx2:
            duplicated_ids[doc_id_1] = {}

    # prin duplicated ids
    print("duplicated", "idx1", index1, "idx2", index2, "num", len(duplicated_ids))


if __name__ == "__main__":
    main()


