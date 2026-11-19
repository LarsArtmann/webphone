package cqrs

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

var _ = Describe("A call-note slot", func() {
	var (
		s     *Notes
		owner domain.Extension
		ctx   context.Context
	)

	BeforeEach(func() {
		s = bootNotes(notesSuiteT, notesSuiteT.TempDir())
		owner = mustExt(notesSuiteT, "1001")
		ctx = context.Background()
	})

	When("first touched by a disposition", func() {
		It("mints the slot with the verdict and an empty body", func() {
			Expect(s.SetDisposition(ctx, owner, "call-1", domain.DispositionVoicemail)).To(Succeed())
			view := awaitNote(notesSuiteT, s, owner, "call-1")
			Expect(view.Disposition).To(Equal(domain.DispositionVoicemail.String()))
			Expect(view.Body).To(BeEmpty())
			Expect(s.Count(ctx, owner)).To(Equal(1))
		})
	})

	When("first touched by a note body", func() {
		It("mints the slot and keeps the verdict untouched", func() {
			changed, err := s.Set(ctx, owner, "call-2", "discussed the invoice")
			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeTrue())
			view := awaitNote(notesSuiteT, s, owner, "call-2")
			Expect(view.Body).To(Equal("discussed the invoice"))
			Expect(view.Disposition).To(BeZero())
		})
	})

	When("the operator re-selects the verdict already on record", func() {
		It("is inert — the slot's stamp does not move", func() {
			Expect(s.SetDisposition(ctx, owner, "call-3", domain.DispositionAnswered)).To(Succeed())
			first := awaitNote(notesSuiteT, s, owner, "call-3")

			Expect(s.SetDisposition(ctx, owner, "call-3", domain.DispositionAnswered)).To(Succeed())
			second, err := s.Get(ctx, owner, "call-3")
			Expect(err).NotTo(HaveOccurred())
			Expect(second).NotTo(BeNil())
			// No journal event fired, so the committed view is the same
			// fact — including its UpdatedAt, exactly (not merely within
			// the same second).
			Expect(second.UpdatedAt).To(Equal(first.UpdatedAt))
			Expect(s.Count(ctx, owner)).To(Equal(1))
		})
	})

	When("the operator corrects the verdict", func() {
		It("journals the correction and preserves the prose", func() {
			Expect(s.SetDisposition(ctx, owner, "call-4", domain.DispositionMissed)).To(Succeed())
			changed, err := s.Set(ctx, owner, "call-4", "left a callback promise")
			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeTrue())

			Expect(s.SetDisposition(ctx, owner, "call-4", domain.DispositionAnswered)).To(Succeed())
			view := awaitNote(notesSuiteT, s, owner, "call-4")
			Expect(view.Disposition).To(Equal(domain.DispositionAnswered.String()))
			Expect(view.Body).To(Equal("left a callback promise"))
		})
	})

	When("the autosave echoes the body already on record", func() {
		It("is inert — the changed verdict stays false", func() {
			changed, err := s.Set(ctx, owner, "call-5", "first draft")
			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeTrue())
			awaitNote(notesSuiteT, s, owner, "call-5")

			changed, err = s.Set(ctx, owner, "call-5", "first draft")
			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeFalse())
		})
	})

	When("the slot is deleted", func() {
		It("frees the cap and answers nil reads", func() {
			Expect(s.SetDisposition(ctx, owner, "call-6", domain.DispositionIVR)).To(Succeed())
			awaitNote(notesSuiteT, s, owner, "call-6")

			Expect(s.Delete(ctx, owner, "call-6")).To(Succeed())
			Eventually(func() bool {
				view, err := s.Get(ctx, owner, "call-6")
				return err == nil && view == nil
			}).Should(BeTrue(), "the slot must read nil after the journal folds the delete")
			Expect(s.Count(ctx, owner)).To(Equal(0))

			err := s.Delete(ctx, owner, "call-6")
			Expect(err).To(MatchError(readmodels.ErrNoteMissing))
		})
	})
})
