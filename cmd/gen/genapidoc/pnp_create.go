package genapidoc

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/handlerpnp"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/httptyped"
	"net/http"
	"time"
)

// PnpCreate
// POST /api/v1/pnp
func PnpCreate(ctx context.Context, components openapi3.Components, path map[string]*openapi3.PathItem) {
	const scopedSchemaName = "PnpCreate"
	const routeName = "Add Push Notification Provider"
	const pathRoute = "/api/v1/pnp"

	// --- Request schema
	allReqSchema := make([]*openapi3.SchemaRef, 0)
	for _, example := range backend.MuxBackend().Examples(ctx) {
		reqStruct := handlerpnp.CreateReq{
			BackendConfig: struct {
				Provider       string      `json:"provider" validate:"required"`
				Label          string      `json:"label" validate:"required"`
				CredentialJSON interface{} `json:"credential_json" validate:"required"`
			}{
				Provider:       example.Provider,
				Label:          fmt.Sprintf("my-%s-config", example.Provider),
				CredentialJSON: example.BackendConfig,
			},
		}

		// generate request
		outReq := MustNewSchemaGenerator(ctx, fmt.Sprintf("%s.%s.", scopedSchemaName, example.Provider), reqStruct)
		for s, ref := range outReq.Schemas {
			components.Schemas[s] = ref
		}

		allReqSchema = append(allReqSchema, &openapi3.SchemaRef{
			Ref: fmt.Sprintf("#/components/schemas/%s", outReq.ParentSchemaName),
		})
	}

	reqSchemaName := fmt.Sprintf("%sReq", scopedSchemaName)
	components.Schemas[reqSchemaName] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			OneOf: allReqSchema,
		},
	}

	reqBody := openapi3.NewRequestBody()
	reqBody.WithJSONSchemaRef(&openapi3.SchemaRef{
		Ref: fmt.Sprintf("#/components/schemas/%s", reqSchemaName),
	})

	components.RequestBodies[scopedSchemaName] = &openapi3.RequestBodyRef{
		Value: reqBody,
	}

	paramClientID := openapi3.NewQueryParameter("client_id").WithDescription("App client id")
	paramClientID.Schema = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}}
	paramClientID.Example = "myapp"
	paramClientID.Required = true

	// --- Response schema
	respStruct := handlerpnp.CreateResp{
		App: httptyped.AppEntity{
			ID:        123,
			ClientID:  "myapp",
			Name:      "My App",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// generate response and add to components
	resp := respbuilder.Success(ctx, respStruct)
	outResp := MustNewSchemaGenerator(ctx, scopedSchemaName+".Resp201.", resp)
	for s, ref := range outResp.Schemas {
		components.Schemas[s] = ref
	}

	// --- final spec
	op := openapi3.NewOperation()
	op.Tags = []string{"PNP"}
	op.Summary = routeName
	op.Description = "Register new Push Notification Provider"
	op.OperationID = scopedSchemaName
	op.AddParameter(paramClientID)

	op.RequestBody = &openapi3.RequestBodyRef{
		Ref: fmt.Sprintf("#/components/requestBodies/%s", scopedSchemaName), // refer to generated name we define above
	}
	op.AddResponse(http.StatusCreated, openapi3.NewResponse().WithJSONSchemaRef(
		&openapi3.SchemaRef{
			Ref: fmt.Sprintf("#/components/schemas/%s", outResp.ParentSchemaName),
		},
	).WithDescription("desc"))

	_, exist := path[pathRoute]
	if !exist {
		path[pathRoute] = &openapi3.PathItem{}
	}

	path[pathRoute].Post = op
}
