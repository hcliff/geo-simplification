package geosimplification_test

import (
	"testing"

	"github.com/golang/geo/s2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	geosimplification "gitlab.com/hcliff/geo-simplification"
	"gitlab.com/hcliff/geo-simplification/internal"
)

func TestSimplify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simplify Suite")
}

var _ = Describe("Simplification unit tests", func() {

	It("should not reduce complexity where none exists", func() {
		input := s2.PolylineFromLatLngs([]s2.LatLng{
			s2.LatLngFromDegrees(0, 0),
			s2.LatLngFromDegrees(1, 1),
			s2.LatLngFromDegrees(2, 2),
			s2.LatLngFromDegrees(3, 3),
		})
		simplified, err := geosimplification.SimplifyLine(*input, 0, 0, true)
		Ω(err).Should(BeNil())
		Ω(simplified).Should(Equal(*input))
	})

	// TODO: get a negative test confirmed for this test.
	It("should ensure that loops do not self intersect", func() {
		// H.C: remember that loops start and begin at the same point
		// if you forget this "closing" the loop can result in an intersection
		// - closing the loop defined as the edge between the penumltimate
		// vertex and the first/last vertex
		input := *s2.PolylineFromLatLngs([]s2.LatLng{
			s2.LatLngFromDegrees(43.023790000000005, -76.4486788),
			s2.LatLngFromDegrees(43.0233744, -76.44862240000002),
			s2.LatLngFromDegrees(43.022486900000004, -76.45023590000001),
			s2.LatLngFromDegrees(43.02233710000001, -76.45022560000002),
			s2.LatLngFromDegrees(43.02226590000001, -76.45063540000001),
			s2.LatLngFromDegrees(43.022119900000014, -76.45087099999999),
			s2.LatLngFromDegrees(43.022221, -76.4509325),
			s2.LatLngFromDegrees(43.0218166, -76.45283279999998),
			s2.LatLngFromDegrees(43.022172300000015, -76.4528584),
			s2.LatLngFromDegrees(43.022603000000004, -76.4507891),
		})

		loop := s2.LoopFromPoints(input)
		simplifiedLoop, err := geosimplification.SimplifyLoop(loop, 0.00000000001, 0, true)
		Ω(err).Should(BeNil())
		Ω(simplifiedLoop.NumVertices()).Should(BeNumerically(">", 2))
		Ω(simplifiedLoop.Validate()).ShouldNot(HaveOccurred())
		// this would/should fail
		resorted := append(simplifiedLoop.Vertices()[1:], simplifiedLoop.Vertices()[0])
		Ω(internal.PolylineSelfIntersects(resorted)).Should(BeNil())
	})

	Context("given a shape that might self intersect", func() {
		input := s2.PolylineFromLatLngs([]s2.LatLng{
			s2.LatLngFromDegrees(45.03008967256179, -85.63249468803406),
			s2.LatLngFromDegrees(45.02955889877115, -85.6320869922638),
			s2.LatLngFromDegrees(45.02903570264613, -85.63207626342773),
			s2.LatLngFromDegrees(45.02902053746971, -85.63342809677124),
			s2.LatLngFromDegrees(45.02953615121298, -85.63219428062439),
			s2.LatLngFromDegrees(45.02990011105872, -85.63337445259094),
		})

		threshold := 0.00000000005

		// not a true test, just assert that given this threshold
		// and no attempt to avoid an intersection one would arise
		BeforeEach(func() {
			simplified, err := geosimplification.SimplifyLine(*input, threshold, 0, false)
			Ω(err).Should(BeNil())
			Ω(len(simplified)).Should(BeNumerically("<", len(*input)))
			Ω(internal.PolylineSelfIntersects(simplified)).ShouldNot(BeNil())
		})

		It("should avoid intersections when simplifying", func() {
			simplified, err := geosimplification.SimplifyLine(*input, threshold, 0, true)
			Ω(err).Should(BeNil())
			Ω(internal.PolylineSelfIntersects(simplified)).Should(BeNil())
		})
	})

	Context("given a minimum point count", func() {
		input := s2.PolylineFromLatLngs([]s2.LatLng{
			s2.LatLngFromDegrees(45.034455200000004, -85.62582019999999),
			s2.LatLngFromDegrees(45.03482089999999, -85.6263255),
			s2.LatLngFromDegrees(45.036493099999994, -85.6278167),
			s2.LatLngFromDegrees(45.036684699999995, -85.62817409999998),
			s2.LatLngFromDegrees(45.036789199999994, -85.62888889999998),
			s2.LatLngFromDegrees(45.036954699999995, -85.6302076),
			s2.LatLngFromDegrees(45.03697210000001, -85.631403),
			s2.LatLngFromDegrees(45.03682090000001, -85.6362185),
			s2.LatLngFromDegrees(45.0352695, -85.6362673),
			s2.LatLngFromDegrees(45.03524919999999, -85.62878779999997),
			s2.LatLngFromDegrees(45.03406769999999, -85.62880649999998),
			s2.LatLngFromDegrees(45.03408939999998, -85.62549969999999),
		})

		// pretty high threshold
		threshold := 0.0001
		minPointsToKeep := 3

		// not a true test, just assert that given this threshold
		// and no attempt to avoid an intersection one would arise
		BeforeEach(func() {
			simplified, err := geosimplification.SimplifyLine(*input, threshold, 0, false)
			Ω(err).Should(BeNil())
			Ω(len(simplified)).Should(BeNumerically("<", minPointsToKeep))
		})

		It("Should preserve the minimum number of points for lines", func() {
			simplified, err := geosimplification.SimplifyLine(*input, threshold, minPointsToKeep, false)
			Ω(err).Should(BeNil())
			Ω(simplified).Should(HaveLen(minPointsToKeep))
		})
	})

})
