package lib

import (
	"encoding/hex"

	git "github.com/libgit2/git2go"
)

// Manifest represents a collection modules in the repository.
type Manifest struct {
	Dir     string
	Sha     string
	Modules Modules
}

// ManifestByPr returns the manifest of pull request.
func ManifestByPr(dir, src, dst string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	srcC, err := getBranchCommit(repo, src)
	if err != nil {
		return nil, err
	}

	dstC, err := getBranchCommit(repo, dst)
	if err != err {
		return nil, err
	}

	a, err := modulesInDiffWithDependents(repo, srcC, dstC)
	if err != nil {
		return nil, err
	}

	return &Manifest{Modules: a, Dir: dir, Sha: srcC.Id().String()}, nil
}

// ManifestBySha returns the manifest as of the specified commit sha.
func ManifestBySha(dir, sha string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	bytes, err := hex.DecodeString(sha)
	if err != nil {
		return nil, wrap(err)
	}

	oid := git.NewOidFromBytes(bytes)
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, wrap(err)
	}

	return fromCommit(repo, dir, commit)
}

// ManifestByBranch returns the manifest as of the tip of the specified branch.
func ManifestByBranch(dir, branch string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	return fromBranch(repo, dir, branch)
}

// ManifestByDiff returns the manifest for the diff between given two commits.
func ManifestByDiff(dir, from, to string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	fromOid, err := git.NewOid(from)
	if err != nil {
		return nil, wrap(err)
	}

	toOid, err := git.NewOid(to)
	if err != nil {
		return nil, wrap(err)
	}

	fromC, err := repo.LookupCommit(fromOid)
	if err != nil {
		return nil, wrap(err)
	}

	toC, err := repo.LookupCommit(toOid)
	if err != nil {
		return nil, wrap(err)
	}

	a, err := modulesInDiffWithDependents(repo, toC, fromC)
	if err != nil {
		return nil, err
	}

	return &Manifest{Modules: a, Dir: dir, Sha: to}, nil
}

func (m *Manifest) indexByName() map[string]*Module {
	return m.Modules.indexByName()
}

func (m *Manifest) indexByPath() map[string]*Module {
	return m.Modules.indexByPath()
}

func fromCommit(repo *git.Repository, dir string, commit *git.Commit) (*Manifest, error) {
	metadataSet, err := discoverMetadata(repo, commit)
	if err != nil {
		return nil, err
	}

	vmods, err := metadataSet.toModules()
	if err != nil {
		return nil, err
	}

	return &Manifest{dir, commit.Id().String(), vmods}, nil
}

func newEmptyManifest(dir string) *Manifest {
	return &Manifest{Modules: []*Module{}, Dir: dir, Sha: ""}
}

func fromBranch(repo *git.Repository, dir string, branch string) (*Manifest, error) {
	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return nil, err
	}

	return fromCommit(repo, dir, commit)
}

func openRepo(dir string) (*git.Repository, *Manifest, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, nil, wrap(err)
	}
	empty, err := repo.IsEmpty()
	if err != nil {
		return nil, nil, wrap(err)
	}

	if empty {
		return nil, newEmptyManifest(dir), nil
	}

	return repo, nil, nil
}
