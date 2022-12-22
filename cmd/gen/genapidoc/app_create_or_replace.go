package genapidoc

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/handlerapp"
	httptyped2 "github.com/yusufsyaifudin/ngendika/transport/restapi/httptyped"
	"net/http"
	"time"
)

// AppCreateOrReplace
// PUT /api/v1/apps/{client_id}
func AppCreateOrReplace(ctx context.Context, components openapi3.Components, path map[string]*openapi3.PathItem) {
	const scopedSchemaName = "AppCreateOrReplace"
	const routeName = "Create or Replace Existing Application"
	const pathRoute = "/api/v1/apps/{client_id}"

	// --- Request schema
	reqStruct := handlerapp.PutAppRequest{
		Name:    "My App",
		Enabled: true,
	}

	// generate request
	outReq := MustNewSchemaGenerator(ctx, scopedSchemaName+".", reqStruct)
	reqSchemaName := outReq.ParentSchemaName
	for s, ref := range outReq.Schemas {
		components.Schemas[s] = ref
	}

	reqBody := openapi3.NewRequestBody()
	reqBody.WithJSONSchemaRef(&openapi3.SchemaRef{
		Ref: fmt.Sprintf("#/components/schemas/%s", reqSchemaName),
	})

	components.RequestBodies[scopedSchemaName] = &openapi3.RequestBodyRef{
		Value: reqBody,
	}

	// --- Response schema
	respStruct := handlerapp.PutAppResp{
		App: httptyped2.AppEntity{
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

	// --- params
	paramAppClientID := openapi3.NewPathParameter("client_id").WithDescription("AppRepo Client ID")
	paramAppClientID.Schema = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}}
	paramAppClientID.Example = "myapp"

	// --- final spec
	op := openapi3.NewOperation()
	op.Tags = []string{"Application"}
	op.Summary = routeName
	op.OperationID = scopedSchemaName
	op.AddParameter(paramAppClientID)

	op.RequestBody = &openapi3.RequestBodyRef{
		Ref: fmt.Sprintf("#/components/requestBodies/%s", scopedSchemaName), // refer to generated name we define above
	}
	op.AddResponse(http.StatusOK, openapi3.NewResponse().WithJSONSchemaRef(
		&openapi3.SchemaRef{
			Ref: fmt.Sprintf("#/components/schemas/%s", outResp.ParentSchemaName),
		},
	).WithDescription("desc"))

	_, exist := path[pathRoute]
	if !exist {
		path[pathRoute] = &openapi3.PathItem{}
	}

	path[pathRoute].Put = op
}
