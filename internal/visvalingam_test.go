package internal_test

import (
	"math"

	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s2"
	"github.com/hcliff/geo-simplification/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func buildRtree(pointList internal.VertexCollection) (*rtreego.Rtree, error) {
	rtree := rtreego.NewTree(3, 1, 5)
	err := pointList.Do(func(point *internal.PointWithTriangle) (err error) {
		point.BBox, err = internal.TriangleBbox(point)
		if err != nil {
			return err
		}
		rtree.Insert(point)
		return nil
	})
	return rtree, err
}

type intersectionTestCase struct {
	Points   []s2.Point
	Element  int
	Expected []s2.Edge
}

var _ = Describe("Visvalingam unit tests", func() {

	Describe("identifying resulting intersections", func() {

		intersectingLatLngs := []s2.Point{
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.264856517201856, -73.32550048828125)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.319325896602095, -73.14971923828125)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.32141999593439, -73.31451416015625)),
			// removing this element would result in a 4 shape
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.2313150803688, -73.4271240234375)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.24179856487036, -73.16619873046875)),
		}

		testCases := []intersectionTestCase{
			{
				Points: []s2.Point{
					s2.PointFromLatLng(s2.LatLngFromDegrees(40.264856517201856, -73.32550048828125)),
					// removing this element shouldn't matter
					s2.PointFromLatLng(s2.LatLngFromDegrees(40.319325896602095, -73.14971923828125)),
					s2.PointFromLatLng(s2.LatLngFromDegrees(40.32141999593439, -73.31451416015625)),
					s2.PointFromLatLng(s2.LatLngFromDegrees(40.2313150803688, -73.4271240234375)),
					s2.PointFromLatLng(s2.LatLngFromDegrees(40.24179856487036, -73.16619873046875)),
				},
				Element:  1,
				Expected: nil,
			},
			{
				Points:  intersectingLatLngs,
				Element: 3,
				Expected: []s2.Edge{
					{intersectingLatLngs[2], intersectingLatLngs[4]},
					{intersectingLatLngs[0], intersectingLatLngs[1]},
				},
			},
		}

		It("Should match the expected results", func() {
			for _, testCase := range testCases {

				pointList := internal.NewPointWithTriangleList()
				points := make([]*internal.PointWithTriangle, len(testCase.Points))
				for i, s2Point := range testCase.Points {
					point := internal.NewPointWithTriangle(s2Point)
					pointList.PushBack(point)
					points[i] = point
				}

				rtree, err := buildRtree(pointList)
				Ω(err).Should(BeNil())

				intersections := internal.CreatesIntersection(rtree, points[testCase.Element])
				Ω(intersections).Should(Equal(testCase.Expected))
			}
		})
	})

	Describe("Determining triangle size from the linked list", func() {
		var pointList *internal.PointWithTriangleList
		var points []*internal.PointWithTriangle
		BeforeEach(func() {
			latLngs := []s2.LatLng{
				s2.LatLngFromDegrees(-85.1842975616455, 45.35054681476437),
				s2.LatLngFromDegrees(-85.17889022827148, 45.34840544954469),
				s2.LatLngFromDegrees(-85.17399787902832, 45.350607133738215),
				s2.LatLngFromDegrees(-85.16829013824463, 45.34879753655915),
			}
			pointList = internal.NewPointWithTriangleList()
			for _, latLng := range latLngs {
				point := internal.NewPointWithTriangle(s2.PointFromLatLng(latLng))
				pointList.PushBack(point)
				points = append(points, point)
			}
		})

		It("Should have inf area on head and tail", func() {
			Ω(internal.TriangleArea(points[0])).Should(Equal(math.Inf(1)))
			Ω(internal.TriangleArea(points[len(points)-1])).Should(Equal(math.Inf(1)))
		})

		It("Should have non zero areas for points in the middle", func() {
			Ω(internal.TriangleArea(points[1])).Should(BeNumerically(">", 0))
			Ω(internal.TriangleArea(points[2])).Should(BeNumerically(">", 0))
		})

	})

})
