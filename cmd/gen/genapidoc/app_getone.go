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

// AppGetOne
// GET /api/v1/apps/{client_id}
func AppGetOne(ctx context.Context, components openapi3.Components, path map[string]*openapi3.PathItem) {
	const scopedSchemaName = "AppGetOne"
	const routeName = "Get Application"
	const pathRoute = "/api/v1/apps/{client_id}"

	// --- Response schema
	respStruct := handlerapp.GetAppByClientIDResp{
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

	path[pathRoute].Get = op
}
