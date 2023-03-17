package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

//go:embed queries/insert-camera.sql
var rawInsertCameraQuery string
var insertCameraQuery string = strings.TrimSpace(rawInsertCameraQuery)

func (s *Store) AddCamera(ctx context.Context, c Camera) (cameraID CameraID, err error) {
	return executeComponentInsert[CameraID](ctx, insertCameraQuery, c, s.db)
}

//go:embed queries/update-camera.sql
var rawUpdateCameraQuery string
var updateCameraQuery string = strings.TrimSpace(rawUpdateCameraQuery)

func (s *Store) UpdateCamera(ctx context.Context, c Camera) (err error) {
	return executeUpdate[CameraID](ctx, updateCameraQuery, c, s.db)
}

//go:embed queries/delete-camera.sql
var rawDeleteCameraQuery string
var deleteCameraQuery string = strings.TrimSpace(rawDeleteCameraQuery)

func (s *Store) DeleteCamera(ctx context.Context, id CameraID) (err error) {
	return executeDelete[CameraID](ctx, deleteCameraQuery, Camera{ID: id}, s.db)
}

//go:embed queries/select-camera.sql
var rawSelectCameraQuery string
var selectCameraQuery string = strings.TrimSpace(rawSelectCameraQuery)

func (s *Store) GetCamera(ctx context.Context, id CameraID) (i Camera, err error) {
	sel := newCamerasSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectCameraQuery, newCameraSelection(id), sel.Step,
	); err != nil {
		return Camera{}, errors.Wrapf(err, "couldn't get camera with id %d", id)
	}
	cameras := sel.Cameras()
	if len(cameras) == 0 {
		return Camera{}, errors.Errorf("couldn't get non-existent camera with id %d", id)
	}
	return cameras[0], nil
}
