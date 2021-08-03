# NNS back

Neural Network Studio API server.

- [API document](https://documenter.getpostman.com/view/13966682/Tzm5GwnJ)

### API List
- Project
  - [x] GET /api/projects
  - [x] GET /api/project/{projectNo}
  - [x] GET /api/project/{projectNo}/config
  - [x] GET /api/project/{projectNo}/content
  - [x] POST /api/project
  - [x] PUT /api/project/{projectNo}/info
  - [x] PUT /api/project/{projectNo}/config
  - [x] PUT /api/project/{projectNo}/content
  - [x] DELETE /api/project/{projectNo}
  - [x] GET /api/project/{projectNo}/code


- Authentication
  - [x] POST /api/login
  - [x] DELETE /api/logout


- User
  - [x] POST /api/user
  - [x] GET /api/user
  - [x] PUT /api/user
  - [x] PUT /api/user/password
  - [x] DELETE /api/user


- Image
  - [x] POST /api/image


- Asset
  - [ ] GET /api/assets
  - [ ] GET /api/asset/info/{assetID}
  - [ ] PUT /api/asset/info/{assetID}
  - [ ] DELETE /api/asset/{assetID}


- Asset Review
  - [ ] GET /api/asset/review
  - [ ] POST /api/asset/review
  - [ ] PUT /api/asset/review/{reviewID}
  - [ ] DELETE /api/asset/review/{reviewID}