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

// AppCreate
// POST /api/v1/apps
func AppCreate(ctx context.Context, components openapi3.Components, path map[string]*openapi3.PathItem) {
	const scopedSchemaName = "AppCreate"
	const routeName = "Create New Application"
	const pathRoute = "/api/v1/apps"

	// --- Request schema
	reqStruct := handlerapp.CreateAppReq{
		Name:     "My App",
		ClientID: "myapp",
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
	respStruct := handlerapp.CreateAppResp{
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

	// --- final spec
	op := openapi3.NewOperation()
	op.Tags = []string{"Application"}
	op.Summary = routeName
	op.Description = "Will create new Application for push notification or sending messages."
	op.OperationID = scopedSchemaName

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
