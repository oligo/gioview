package explorer

import "testing"

func TestFindNodeInTree(t *testing.T) {
	root := &EntryNode{Path: "/a"}
	root.children = append(root.children, &EntryNode{Path: "/a/b", Parent: root})

	ab := root.children[0]
	ab.children = []*EntryNode{
		{Path: "/a/b/c", Parent: ab},
		{Path: "/a/b/d", Parent: ab},
		{Path: "/a/b/e", Parent: ab},
	}

	abe := ab.children[2]
	abe.children = []*EntryNode{
		{Path: "/a/b/e/k", Parent: abe},
		{Path: "/a/b/e/l", Parent: abe},
		{Path: "/a/b/e/m", Parent: abe},
	}

	cases := []struct {
		current *EntryNode
		path    string
		want    string
	}{
		{current: root, path: "/a/b/c", want: "/a/b/c"},
		{current: root.children[0].children[0], path: "/a/b/e/k", want: "/a/b/e/k"},
		{current: root.children[0].children[2], path: "/a/b", want: "/a/b"},
		{current: root.children[0], path: "/out", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			found := findNodeInTree(tc.current, tc.path)
			if found == nil {
				if tc.want != "" {
					t.Fail()
				}
			} else {
				if tc.want != found.Path {
					t.Fail()
				}
			}
		})

	}

}

func TestLongestCommonPath(t *testing.T) {
	cases := []struct {
		pair []string
		want string
	}{
		{
			pair: []string{
				"/home/User/Desktop/gfg/test",
				"/home/User/Desktop/gfg/file",
				"/home/User/Desktop/geeks/folders",
			},
			want: "/home/User/Desktop/gfg",
		},
		{
			pair: []string{
				"/home/User/Desktop/gfg/test",
				"/home/User/Desktop/gfg/test",
			},
			want: "/home/User/Desktop/gfg/test",
		},
		{
			pair: []string{
				"/home/User/Desktop/abc/test",
				"/home/User/Desktop/gfg/test",
			},
			want: "/home/User/Desktop",
		},
		{
			pair: []string{
				"/home/User/Desktop/abc/test1",
				"/home/User/Desktop/abc/test",
			},
			want: "/home/User/Desktop/abc",
		},
		{
			pair: []string{
				"/home/User/Desktop/abc/test/sudir",
				"/home/User/Desktop/abc/test",
			},
			want: "/home/User/Desktop/abc/test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			cp := longestCommonPath(tc.pair[0], tc.pair[1])
			if cp != tc.want {
				t.Logf("actual: %s, wanted: %s", cp, tc.want)
				t.Fail()
			}
		})
	}

}
