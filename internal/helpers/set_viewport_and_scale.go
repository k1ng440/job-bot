package helpers

// func SetViewportAndScale(w, h int64, scale float64) pcdp.ActionFunc {
// 	return func(ctxt context.Context, ha cdptypes.Handler) error {
// 		err := emulation.SetVisibleSize(w, h).Do(ctxt, ha)
// 		if err != nil {
// 			return err
// 		}
// 		sw, sh := int64(float64(w)/scale), int64(float64(h)/scale)
// 		err = emulation.SetDeviceMetricsOverride(sw, sh, scale, false, false).WithScale(scale).Do(ctxt, ha)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}
// }
