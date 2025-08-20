#!/bin/bash

# Script to clean up a failed release and prepare for retry
# Usage: ./cleanup_release.sh v0.1.7

TAG=${1:-v0.1.7}

echo "This script will help clean up the failed release for $TAG"
echo "You will need to:"
echo ""
echo "1. Go to https://github.com/trahma/certwiz/releases/tag/$TAG"
echo "2. Click 'Delete' to delete the release (but keep the tag)"
echo ""
echo "3. Then run the following commands to delete and recreate the tag:"
echo ""
echo "   git tag -d $TAG"
echo "   git push origin :refs/tags/$TAG"
echo "   git tag -a $TAG -m \"Release $TAG\""
echo "   git push origin $TAG"
echo ""
echo "This will trigger a fresh goreleaser build with no conflicts."