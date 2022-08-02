package export

type exportMetadata struct {
	ActualShardID          uint32 `json:"actualShardID"`
	ActualShardHasValue    bool   `json:"actualShardHasValue"`
	ProjectedShardID       uint32 `json:"projectedShardID"`
	ProjectedShardHasValue bool   `json:"projectedShardHasValue"`
	StartIndex             int    `json:"startIndex"`
	NumKeys                int    `json:"numKeys"`
}
