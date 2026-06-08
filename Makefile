
FLOW_V1_SCHEMA := internal/contracts/flowschema/flow_v1.schema.json
FLOW_V1_GEN_OUTPUT_FILE := internal/contracts/flowschema/flow_v1.gen.go
FLOW_PACKAGE := flowschema

flow_v1_codegen:
	go-jsonschema -p $(FLOW_PACKAGE) --schema-root-type=urn:schemas:flow:v1=FlowV1 $(FLOW_V1_SCHEMA) > $(FLOW_V1_GEN_OUTPUT_FILE)

METADATA_V1_SCHEMA := internal/contracts/metadataschema/artifact_metadata_v1.schema.json
METADATA_V1_GEN_OUTPUT_FILE := internal/contracts/metadataschema/artifact_metadata_v1.gen.go
METADATA_PACKAGE := metadataschema

metadata_v1_codegen:
	go-jsonschema -p $(METADATA_PACKAGE) --schema-root-type=urn:schemas:artifact_metadata:v1=ArtifactMetadataV1 $(METADATA_V1_SCHEMA) > $(METADATA_V1_GEN_OUTPUT_FILE)
