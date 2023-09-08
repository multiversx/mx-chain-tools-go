from elasticsearch import Elasticsearch, exceptions as es_exceptions
from elasticsearch.helpers import scan


class ElasticClient:
    def __init__(self, url: str, user: str, password: str):
        self.url = url
        self.username = user
        self.password = password
        self.es_client = self.create_elasticsearch_client()

    def create_elasticsearch_client(self) -> Elasticsearch | None:
        try:
            es = Elasticsearch([self.url], http_auth=(self.username, self.password))
            if es.ping():
                return es
            else:
                raise es_exceptions.ConnectionError("Failed to connect to Elasticsearch.")
        except es_exceptions.ConnectionError as e:
            print(f"Error initializing Elasticsearch client: {str(e)}")
            return None

    def get_indices_behind_alias(self, alias: str) -> [str]:
        try:
            indices = self.es_client.indices.get_alias(name=alias)
            return list(indices.keys())
        except Exception as e:
            print(f"Error: {str(e)}")
            return []

    def get_ids_from_index(self, index_name: str, num_ids: int, sort: str):
        search_query = {
            "query": {
                "match_all": {}
            },
            "sort": [{"timestamp": sort}],
            "size": num_ids
        }
        try:
            results = scan(
                self.es_client,
                index=index_name,
                query=search_query,
                size=9999,
                scroll="5m"  # Scroll time for fetching documents
            )
            ids = {}
            for result in results:
                ids[result["_id"]] = {}
                if len(ids) > num_ids:
                    break
            return ids
        except Exception as e:
            print(f"Error: {str(e)}")
            return {}
