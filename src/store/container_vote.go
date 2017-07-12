package store

import (
	"github.com/pkg/errors"
	"github.com/skycoin/bbs/src/store/typ"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
)

// GetVotesForThread obtains the votes for specified thread present in specified board.
func (c *CXO) GetVotesForThread(bpk cipher.PubKey, tRef skyobject.Reference) ([]typ.Vote, error) {
	c.Lock(c.GetVotesForThread)
	defer c.Unlock()
	return c.ss.GetThreadVotes(bpk, tRef)
}

// GetVotesForPost obtains the votes for specified post present in specified board.
func (c *CXO) GetVotesForPost(bpk cipher.PubKey, pRef skyobject.Reference) ([]typ.Vote, error) {
	c.Lock(c.GetVotesForPost)
	defer c.Unlock()
	return c.ss.GetPostVotes(bpk, pRef)
}

// VoteForThread adds a vote for a thread on a specified board.
func (c *CXO) AddVoteForThread(bpk cipher.PubKey, bsk cipher.SecKey, tRef skyobject.Reference, newVote *typ.Vote) error {
	c.Lock(c.AddVoteForThread)
	defer c.Unlock()

	w := c.c.LastRootSk(bpk, bsk).Walker()
	vc := &typ.ThreadVotesContainer{}
	if e := w.AdvanceFromRoot(vc, makeThreadVotesContainerFinder(w.Root())); e != nil {
		return e
	}
	// Obtain vote references.
	voteRefs, e := vc.GetThreadVotes(tRef)
	if e != nil {
		return e
	}
	// Loop through votes to see if user has already voted.
	for i, vRef := range voteRefs.Votes {
		tempVote := &typ.Vote{}
		if e := w.DeserializeFromRef(vRef, tempVote); e != nil {
			return errors.Wrap(e, "failed to obtain vote")
		}
		// Replace vote if already voted.
		if tempVote.User == newVote.User {
			voteRefs.Votes[i] = w.Root().Save(*newVote)
			goto SaveThreadVotesContainer
		}
	}
	// If user has not already voted - add a new vote.
	voteRefs.Votes = append(voteRefs.Votes, w.Root().Save(*newVote))

SaveThreadVotesContainer:
	c.ss.Fill(w.Root()) // TODO: Optimise.
	return w.ReplaceCurrent(*vc)
}

// VoteForPost adds a vote for a post on a specified board.
func (c *CXO) AddVoteForPost(bpk cipher.PubKey, bsk cipher.SecKey, pRef skyobject.Reference, newVote *typ.Vote) error {
	c.Lock(c.AddVoteForPost)
	defer c.Unlock()

	w := c.c.LastRootSk(bpk, bsk).Walker()
	vc := &typ.PostVotesContainer{}
	if e := w.AdvanceFromRoot(vc, makePostVotesContainerFinder(w.Root())); e != nil {
		return e
	}
	// Obtain vote references.
	voteRefs, e := vc.GetPostVotes(pRef)
	if e != nil {
		return e
	}
	// Loop through votes to see if user has voted already.
	for i, vRef := range voteRefs.Votes {
		tempVote := &typ.Vote{}
		if e := w.DeserializeFromRef(vRef, tempVote); e != nil {
			return errors.Wrap(e, "failed to obtain vote")
		}
		// Replace vote if already voted.
		if tempVote.User == newVote.User {
			voteRefs.Votes[i] = w.Root().Save(*newVote)
			goto SavePostVotesContainer
		}
	}
	// If user has not already voted - add a new vote.
	voteRefs.Votes = append(voteRefs.Votes, w.Root().Save(*newVote))

SavePostVotesContainer:
	c.ss.Fill(w.Root()) // TODO: Optimise.
	return w.ReplaceCurrent(*vc)
}

// RemoveVoteForThread removes a vote completely for a thread and specified user.
func (c *CXO) RemoveVoteForThread(upk, bpk cipher.PubKey, bsk cipher.SecKey, tRef skyobject.Reference) error {
	c.Lock(c.RemoveVoteForThread)
	defer c.Unlock()

	w := c.c.LastRootSk(bpk, bsk).Walker()
	vc := &typ.ThreadVotesContainer{}
	if e := w.AdvanceFromRoot(vc, makeThreadVotesContainerFinder(w.Root())); e != nil {
		return e
	}
	// Obtain vote references.
	voteRefs, e := vc.GetThreadVotes(tRef)
	if e != nil {
		return errors.Wrap(e, "failed to remove vote for thread")
	}
	// loop through votes to see what to remove.
	for i, vRef := range voteRefs.Votes {
		tempVote := &typ.Vote{}
		if e := w.DeserializeFromRef(vRef, tempVote); e != nil {
			return errors.Wrap(e, "failed to obtain vote")
		}
		if tempVote.User == upk {
			// Delete of index i.
			voteRefs.Votes[i], voteRefs.Votes[len(voteRefs.Votes)-1] =
				voteRefs.Votes[len(voteRefs.Votes)-1], voteRefs.Votes[i]
			voteRefs.Votes = voteRefs.Votes[:len(voteRefs.Votes)-1]
			// MemoryMode.
			c.ss.Fill(w.Root()) // TODO: Optimise.
			return w.ReplaceCurrent(*vc)
		}
	}
	c.ss.Fill(w.Root()) // TODO: Optimise.
	return nil
}

// RemoveVoteForPost removes a vote completely for a post and specified user.
func (c *CXO) RemoveVoteForPost(upk, bpk cipher.PubKey, bsk cipher.SecKey, pRef skyobject.Reference) error {
	c.Lock(c.RemoveVoteForPost)
	defer c.Unlock()

	w := c.c.LastRootSk(bpk, bsk).Walker()
	vc := &typ.PostVotesContainer{}
	if e := w.AdvanceFromRoot(vc, makePostVotesContainerFinder(w.Root())); e != nil {
		return e
	}
	// Obtain vote references.
	voteRefs, e := vc.GetPostVotes(pRef)
	if e != nil {
		return e
	}
	// loop through votes to see what to remove.
	for i, vRef := range voteRefs.Votes {
		tempVote := &typ.Vote{}
		if e := w.DeserializeFromRef(vRef, tempVote); e != nil {
			return errors.Wrap(e, "failed to obtain vote")
		}
		if tempVote.User == upk {
			// Delete of index i.
			voteRefs.Votes[i], voteRefs.Votes[len(voteRefs.Votes)-1] =
				voteRefs.Votes[len(voteRefs.Votes)-1], voteRefs.Votes[i]
			voteRefs.Votes = voteRefs.Votes[:len(voteRefs.Votes)-1]
			// MemoryMode.
			c.ss.Fill(w.Root()) // TODO: Optimise.
			return w.ReplaceCurrent(*vc)
		}
	}
	c.ss.Fill(w.Root()) // TODO: Optimise.
	return nil
}
